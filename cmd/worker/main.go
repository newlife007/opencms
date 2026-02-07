package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openwan/media-asset-management/internal/config"
	"github.com/openwan/media-asset-management/internal/queue"
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
	fmt.Println()

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
		go worker(ctx, workerID, queueService, ffmpegWrapper, cfg.FFmpeg.Parameters)
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

func worker(ctx context.Context, workerID int, queueService queue.QueueService, ffmpegWrapper *transcoding.FFmpegWrapper, defaultParams string) {
	queueName := "openwan_transcoding_jobs"

	fmt.Printf("[Worker %d] Started, subscribing to queue: %s\n", workerID, queueName)

	// Subscribe to queue
	err := queueService.Subscribe(ctx, queueName, func(message *queue.Message) error {
		return handleTranscodeJob(workerID, message, ffmpegWrapper, defaultParams)
	})

	if err != nil {
		log.Printf("[Worker %d] Error: %v\n", workerID, err)
	}
}

func handleTranscodeJob(workerID int, message *queue.Message, ffmpegWrapper *transcoding.FFmpegWrapper, defaultParams string) error {
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

	// Execute transcoding
	startTime := time.Now()
	
	// For S3 storage, we would need to download first
	// For now, assume local access or that path is accessible
	opts := transcoding.TranscodeOptions{
		InputPath:    job.InputPath,
		OutputPath:   job.OutputPath,
		CustomParams: params,
	}
	
	err := ffmpegWrapper.Transcode(context.Background(), opts)
	
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[Worker %d] ✗ Transcoding failed (%.2fs): %v\n", workerID, duration.Seconds(), err)
		return err
	}

	fmt.Printf("[Worker %d] ✓ Transcoding completed (%.2fs)\n", workerID, duration.Seconds())
	fmt.Printf("[Worker %d]   Preview file: %s\n", workerID, job.OutputPath)

	return nil
}
