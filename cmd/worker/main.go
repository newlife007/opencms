package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/openwan/media-asset-management/internal/config"
	"github.com/openwan/media-asset-management/internal/queue"
	"github.com/openwan/media-asset-management/internal/storage"
	"github.com/openwan/media-asset-management/internal/transcoding"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("OpenWan Transcoding Worker")
	fmt.Println("Version: 1.0.0")
	fmt.Println("========================================")
	fmt.Println()

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("✓ Configuration loaded\n")
	fmt.Printf("  FFmpeg: %s\n", cfg.FFmpeg.BinaryPath)
	fmt.Printf("  Queue: %s\n", cfg.Queue.RabbitMQURL)
	fmt.Printf("  Workers: %d\n", cfg.FFmpeg.WorkerCount)
	fmt.Printf("  Storage: %s\n", cfg.Storage.Type)
	if cfg.Storage.Type == "s3" {
		fmt.Printf("  S3 Bucket: %s\n", cfg.Storage.S3Bucket)
		fmt.Printf("  S3 Region: %s\n", cfg.Storage.S3Region)
	}
	fmt.Println()

	// Initialize Storage service
	storageConfig := storage.Config{
		Type:         cfg.Storage.Type,
		LocalPath:    cfg.Storage.LocalPath,
		S3Bucket:     cfg.Storage.S3Bucket,
		S3Region:     cfg.Storage.S3Region,
		S3Prefix:     cfg.Storage.S3Prefix,
		S3UseIAMRole: true, // Use IAM role for EC2 instance
	}
	storageService, err := storage.NewStorageFromConfig(storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	fmt.Println("✓ Storage service initialized")

	// Initialize FFmpeg service
	ffmpegWrapper := transcoding.NewFFmpegWrapper(cfg.FFmpeg.BinaryPath, cfg.FFmpeg.Parameters)
	fmt.Println("✓ FFmpeg service initialized")

	// Initialize queue connection
	queueService, err := queue.NewRabbitMQQueue(cfg.Queue.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to queue: %v", err)
	}
	defer queueService.Close()
	fmt.Println("✓ Connected to message queue")
	fmt.Println()

	// Create worker pool
	workerCount := cfg.FFmpeg.WorkerCount
	if workerCount <= 0 {
		workerCount = 2
	}

	fmt.Printf("Starting %d workers...\n", workerCount)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start workers
	for i := 0; i < workerCount; i++ {
		workerID := i + 1
		go worker(ctx, workerID, queueService, ffmpegWrapper, storageService, cfg.FFmpeg.Parameters)
	}

	fmt.Printf("✓ All workers started\n")
	fmt.Println("\n========================================")
	fmt.Println("Worker service is running")
	fmt.Println("Waiting for transcoding jobs...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("========================================")
	fmt.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n\nReceived shutdown signal, stopping workers...")
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Println("✓ Worker service stopped")
}

func worker(ctx context.Context, workerID int, queueService queue.QueueService, ffmpegWrapper *transcoding.FFmpegWrapper, storageService storage.StorageService, defaultParams string) {
	queueName := "openwan_transcoding_jobs"

	fmt.Printf("[Worker %d] Started, subscribing to queue: %s\n", workerID, queueName)

	// Subscribe to queue
	err := queueService.Subscribe(ctx, queueName, func(message *queue.Message) error {
		return handleTranscodeJob(workerID, message, ffmpegWrapper, storageService, defaultParams)
	})

	if err != nil {
		log.Printf("[Worker %d] Error: %v\n", workerID, err)
	}
}

func handleTranscodeJob(workerID int, message *queue.Message, ffmpegWrapper *transcoding.FFmpegWrapper, storageService storage.StorageService, defaultParams string) error {
	// Parse job data
	var job queue.TranscodeJob
	if err := json.Unmarshal([]byte(message.Body), &job); err != nil {
		log.Printf("[Worker %d] Failed to parse job: %v\n", workerID, err)
		return err
	}

	fmt.Printf("[Worker %d] Processing job for file %d\n", workerID, job.FileID)
	fmt.Printf("[Worker %d]   Input: %s\n", workerID, job.InputPath)
	fmt.Printf("[Worker %d]   Output: %s\n", workerID, job.OutputPath)
	fmt.Printf("[Worker %d]   Storage: %s\n", workerID, job.StorageType)

	// Use job parameters or default
	params := job.Parameters
	if params == "" {
		params = defaultParams
	}

	var inputFile, outputFile string
	var cleanupFiles []string
	
	// For S3 storage, download input file to temp directory
	if job.StorageType == "s3" {
		tempDir := "/tmp/openwan-transcode"
		os.MkdirAll(tempDir, 0755)
		
		// Download input file
		inputExt := filepath.Ext(job.InputPath)
		inputFile = filepath.Join(tempDir, fmt.Sprintf("input-%d-%d%s", job.FileID, time.Now().Unix(), inputExt))
		cleanupFiles = append(cleanupFiles, inputFile)
		
		fmt.Printf("[Worker %d]   ⬇  Downloading from S3: %s\n", workerID, job.InputPath)
		
		reader, err := storageService.Download(context.Background(), job.InputPath)
		if err != nil {
			log.Printf("[Worker %d] ✗ Failed to download from S3: %v\n", workerID, err)
			return fmt.Errorf("failed to download input file: %w", err)
		}
		defer reader.Close()
		
		// Write to temp file
		f, err := os.Create(inputFile)
		if err != nil {
			log.Printf("[Worker %d] ✗ Failed to create temp input file: %v\n", workerID, err)
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		
		written, err := io.Copy(f, reader)
		f.Close()
		if err != nil {
			log.Printf("[Worker %d] ✗ Failed to write temp input file: %v\n", workerID, err)
			os.Remove(inputFile)
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		
		fmt.Printf("[Worker %d]   ✓ Downloaded %.2f MB to %s\n", workerID, float64(written)/(1024*1024), inputFile)
		
		// Set output file path
		outputFile = filepath.Join(tempDir, fmt.Sprintf("output-%d-%d.flv", job.FileID, time.Now().Unix()))
		cleanupFiles = append(cleanupFiles, outputFile)
	} else {
		// Local storage - use paths directly
		inputFile = job.InputPath
		outputFile = job.OutputPath
	}
	
	// Cleanup temp files on exit
	defer func() {
		for _, f := range cleanupFiles {
			os.Remove(f)
		}
	}()

	// Execute transcoding
	startTime := time.Now()
	fmt.Printf("[Worker %d]   🎥 Transcoding: %s -> %s\n", workerID, inputFile, outputFile)
	
	opts := transcoding.TranscodeOptions{
		InputPath:    inputFile,
		OutputPath:   outputFile,
		CustomParams: params,
	}
	
	err := ffmpegWrapper.Transcode(context.Background(), opts)
	
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[Worker %d] ✗ Transcoding failed (%.2fs): %v\n", workerID, duration.Seconds(), err)
		return err
	}

	fmt.Printf("[Worker %d] ✓ Transcoding completed (%.2fs)\n", workerID, duration.Seconds())
	
	// For S3 storage, upload output file
	if job.StorageType == "s3" {
		fmt.Printf("[Worker %d]   ⬆  Uploading to S3: %s\n", workerID, job.OutputPath)
		
		f, err := os.Open(outputFile)
		if err != nil {
			log.Printf("[Worker %d] ✗ Failed to open output file: %v\n", workerID, err)
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer f.Close()
		
		// Get file size
		stat, _ := f.Stat()
		outputSize := stat.Size()
		
		metadata := map[string]string{
			"Content-Type":   "video/x-flv",
			"original-file":  fmt.Sprintf("%d", job.FileID),
			"transcode-date": time.Now().Format(time.RFC3339),
		}
		
		uploadedPath, err := storageService.Upload(context.Background(), job.OutputPath, f, metadata)
		if err != nil {
			log.Printf("[Worker %d] ✗ Failed to upload to S3: %v\n", workerID, err)
			return fmt.Errorf("failed to upload output file: %w", err)
		}
		
		fmt.Printf("[Worker %d]   ✓ Uploaded %.2f MB to S3: %s\n", workerID, float64(outputSize)/(1024*1024), uploadedPath)
	}
	
	fmt.Printf("[Worker %d] ✅ Job completed for file %d\n", workerID, job.FileID)

	return nil
}
