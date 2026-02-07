package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openwan/media-asset-management/internal/queue"
	"github.com/openwan/media-asset-management/internal/transcoding"
)

// TranscodingWorker processes transcoding jobs from queue
type TranscodingWorker struct {
	queue            queue.QueueService
	transcodeService *transcoding.TranscodeService
	workerID         string
}

// NewTranscodingWorker creates a new transcoding worker
func NewTranscodingWorker(q queue.QueueService, ts *transcoding.TranscodeService, workerID string) *TranscodingWorker {
	return &TranscodingWorker{
		queue:            q,
		transcodeService: ts,
		workerID:         workerID,
	}
}

// Start starts the worker to consume and process jobs
func (w *TranscodingWorker) Start(ctx context.Context) error {
	return w.queue.Subscribe(ctx, "transcode_jobs", w.processJob)
}

// processJob processes a single transcoding job
func (w *TranscodingWorker) processJob(msg *queue.Message) error {
	var job queue.TranscodeJob
	if err := json.Unmarshal([]byte(msg.Body), &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}
	
	fmt.Printf("[Worker %s] Processing transcoding job for file %d\n", w.workerID, job.FileID)
	
	// TODO: Update job status to "processing" in database
	// TODO: Download file from storage if S3
	// TODO: Transcode file
	// TODO: Upload preview to storage
	// TODO: Update job status to "completed" or "failed"
	
	// Simulate processing
	time.Sleep(1 * time.Second)
	
	fmt.Printf("[Worker %s] Completed transcoding job for file %d\n", w.workerID, job.FileID)
	
	return nil
}
