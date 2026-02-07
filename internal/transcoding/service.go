package transcoding

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// TranscodeService handles media transcoding operations
type TranscodeService struct {
	ffmpeg  *FFmpegWrapper
	tempDir string
}

// NewTranscodeService creates a new transcode service
func NewTranscodeService(ffmpegPath, ffmpegParams, tempDir string) *TranscodeService {
	return &TranscodeService{
		ffmpeg:  NewFFmpegWrapper(ffmpegPath, ffmpegParams),
		tempDir: tempDir,
	}
}

// TranscodeForPreview transcodes a file to FLV preview format
func (s *TranscodeService) TranscodeForPreview(ctx context.Context, fileID uint64, inputPath, outputPath string) error {
	// Check if preview already exists
	if _, err := os.Stat(outputPath); err == nil {
		// Preview already exists
		info, _ := os.Stat(outputPath)
		if info.Size() > 0 {
			return nil // Preview file exists and has content
		}
	}
	
	// Create lock to prevent concurrent transcoding
	lockPath := outputPath + ".lock"
	lock := NewFileLock(lockPath)
	
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("transcoding already in progress: %w", err)
	}
	defer lock.Unlock()
	
	// Double-check after acquiring lock
	if _, err := os.Stat(outputPath); err == nil {
		info, _ := os.Stat(outputPath)
		if info.Size() > 0 {
			return nil
		}
	}
	
	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Transcode with progress callback
	progressCallback := func(progress float64) {
		// TODO: Update progress in database or cache
		fmt.Printf("Transcoding file %d: %.2f%%\n", fileID, progress)
	}
	
	opts := TranscodeOptions{
		InputPath:        inputPath,
		OutputPath:       outputPath,
		ProgressCallback: progressCallback,
	}
	
	if err := s.ffmpeg.Transcode(ctx, opts); err != nil {
		return fmt.Errorf("transcoding failed: %w", err)
	}
	
	return nil
}

// GetFFmpegVersion returns FFmpeg version information
func (s *TranscodeService) GetFFmpegVersion(ctx context.Context) (string, error) {
	return s.ffmpeg.GetVersion(ctx)
}
