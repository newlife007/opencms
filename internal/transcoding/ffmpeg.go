package transcoding

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FFmpegWrapper wraps FFmpeg command execution
type FFmpegWrapper struct {
	binaryPath string
	params     string
	timeout    time.Duration
}

// NewFFmpegWrapper creates a new FFmpeg wrapper
func NewFFmpegWrapper(binaryPath, params string) *FFmpegWrapper {
	return &FFmpegWrapper{
		binaryPath: binaryPath,
		params:     params,
		timeout:    3600 * time.Second, // Default 1 hour timeout
	}
}

// SetTimeout sets the timeout for transcoding operations
func (f *FFmpegWrapper) SetTimeout(timeout time.Duration) {
	f.timeout = timeout
}

// TranscodeOptions contains options for transcoding
type TranscodeOptions struct {
	InputPath        string
	OutputPath       string
	CustomParams     string
	ProgressCallback func(progress float64)
}

// Transcode executes FFmpeg transcoding with full error handling
func (f *FFmpegWrapper) Transcode(ctx context.Context, opts TranscodeOptions) error {
	// Validate input file exists
	if _, err := os.Stat(opts.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", opts.InputPath)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(opts.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	// Build FFmpeg command
	args := f.buildFFmpegArgs(opts)

	cmd := exec.CommandContext(timeoutCtx, f.binaryPath, args...)

	// Capture stderr for progress tracking and error messages
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Capture stdout as well
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Parse FFmpeg output for progress in goroutine
	errChan := make(chan error, 1)
	var stderrOutput strings.Builder

	go func() {
		if opts.ProgressCallback != nil {
			errChan <- f.parseProgress(stderr, opts.ProgressCallback, &stderrOutput)
		} else {
			// Just capture stderr without progress tracking
			_, err := io.Copy(&stderrOutput, stderr)
			errChan <- err
		}
	}()

	// Discard stdout
	go io.Copy(io.Discard, stdout)

	// Wait for command to complete
	cmdErr := cmd.Wait()

	// Wait for progress parsing to complete
	parseErr := <-errChan

	// Check for timeout
	if timeoutCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("transcoding timed out after %v", f.timeout)
	}

	// Check for command errors
	if cmdErr != nil {
		stderr := stderrOutput.String()
		if stderr != "" {
			return fmt.Errorf("FFmpeg execution failed: %w\\nStderr: %s", cmdErr, stderr)
		}
		return fmt.Errorf("FFmpeg execution failed: %w", cmdErr)
	}

	// Verify output file was created
	if _, err := os.Stat(opts.OutputPath); os.IsNotExist(err) {
		return fmt.Errorf("output file was not created: %s", opts.OutputPath)
	}

	// Check output file is not empty
	info, err := os.Stat(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to stat output file: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("output file is empty")
	}

	return parseErr
}

// buildFFmpegArgs builds the FFmpeg command arguments
func (f *FFmpegWrapper) buildFFmpegArgs(opts TranscodeOptions) []string {
	args := []string{
		"-i", opts.InputPath,
		"-progress", "pipe:2", // Send progress to stderr
		"-loglevel", "info",   // Set log level
	}

	// Add custom parameters if provided
	if opts.CustomParams != "" {
		paramList := strings.Fields(opts.CustomParams)
		args = append(args, paramList...)
	} else if f.params != "" {
		// Use default parameters
		paramList := strings.Fields(f.params)
		args = append(args, paramList...)
	} else {
		// Fallback to standard FLV parameters
		args = append(args,
			"-y",           // Overwrite output
			"-ab", "56k",   // Audio bitrate
			"-ar", "22050", // Audio sample rate
			"-r", "15",     // Frame rate
			"-b:v", "500k", // Video bitrate
			"-s", "320x240", // Video size
			"-f", "flv",    // Output format
		)
	}

	args = append(args, opts.OutputPath)

	return args
}

// parseProgress parses FFmpeg output to extract progress information
func (f *FFmpegWrapper) parseProgress(stderr io.Reader, callback func(float64), output *strings.Builder) error {
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for long lines

	durationRegex := regexp.MustCompile(`Duration: (\d+):(\d+):(\d+\.?\d*)`)
	timeRegex := regexp.MustCompile(`time=(\d+):(\d+):(\d+\.?\d*)`)

	var totalDuration float64
	var lastProgress float64

	for scanner.Scan() {
		line := scanner.Text()

		// Write to output buffer for error reporting
		if output != nil {
			output.WriteString(line)
			output.WriteString("\\n")
		}

		// Extract total duration
		if matches := durationRegex.FindStringSubmatch(line); len(matches) == 4 {
			hours, _ := strconv.ParseFloat(matches[1], 64)
			minutes, _ := strconv.ParseFloat(matches[2], 64)
			seconds, _ := strconv.ParseFloat(matches[3], 64)
			totalDuration = hours*3600 + minutes*60 + seconds
		}

		// Extract current time and calculate progress
		if matches := timeRegex.FindStringSubmatch(line); len(matches) == 4 && totalDuration > 0 {
			hours, _ := strconv.ParseFloat(matches[1], 64)
			minutes, _ := strconv.ParseFloat(matches[2], 64)
			seconds, _ := strconv.ParseFloat(matches[3], 64)
			currentTime := hours*3600 + minutes*60 + seconds

			progress := (currentTime / totalDuration) * 100
			if progress > 100 {
				progress = 100
			}

			// Only call callback if progress changed significantly (avoid spamming)
			if progress-lastProgress >= 1.0 || progress >= 100 {
				callback(progress)
				lastProgress = progress
			}
		}
	}

	return scanner.Err()
}

// GetVersion returns FFmpeg version information
func (f *FFmpegWrapper) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, f.binaryPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get FFmpeg version: %w", err)
	}
	return string(output), nil
}

// GetMediaInfo extracts media information from a file
func (f *FFmpegWrapper) GetMediaInfo(ctx context.Context, filePath string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, f.binaryPath, "-i", filePath, "-hide_banner")
	output, _ := cmd.CombinedOutput() // FFmpeg writes info to stderr, so error is expected

	info := make(map[string]interface{})
	outputStr := string(output)

	// Extract duration
	durationRegex := regexp.MustCompile(`Duration: (\d+):(\d+):(\d+\.?\d*)`)
	if matches := durationRegex.FindStringSubmatch(outputStr); len(matches) == 4 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		minutes, _ := strconv.ParseFloat(matches[2], 64)
		seconds, _ := strconv.ParseFloat(matches[3], 64)
		info["duration"] = hours*3600 + minutes*60 + seconds
	}

	// Extract video codec
	videoRegex := regexp.MustCompile(`Video: ([^,]+)`)
	if matches := videoRegex.FindStringSubmatch(outputStr); len(matches) == 2 {
		info["video_codec"] = strings.TrimSpace(matches[1])
	}

	// Extract audio codec
	audioRegex := regexp.MustCompile(`Audio: ([^,]+)`)
	if matches := audioRegex.FindStringSubmatch(outputStr); len(matches) == 2 {
		info["audio_codec"] = strings.TrimSpace(matches[1])
	}

	// Extract resolution
	resolutionRegex := regexp.MustCompile(`(\d+)x(\d+)`)
	if matches := resolutionRegex.FindStringSubmatch(outputStr); len(matches) == 3 {
		info["width"] = matches[1]
		info["height"] = matches[2]
	}

	return info, nil
}

// ValidateFFmpeg checks if FFmpeg is available and working
func (f *FFmpegWrapper) ValidateFFmpeg(ctx context.Context) error {
	// Check if binary exists
	if _, err := os.Stat(f.binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("FFmpeg binary not found at: %s", f.binaryPath)
	}

	// Try to get version
	_, err := f.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("FFmpeg validation failed: %w", err)
	}

	return nil
}
