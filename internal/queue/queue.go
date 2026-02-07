package queue

import (
	"context"
	"time"
)

// Message represents a queue message
type Message struct {
	ID         string
	Body       string
	Attributes map[string]string
	Timestamp  time.Time
	Attempts   int
}

// QueueService defines the message queue interface
type QueueService interface {
	Publish(ctx context.Context, queueName string, message *Message) error
	Subscribe(ctx context.Context, queueName string, handler func(*Message) error) error
	Close() error
}

// JobType defines types of background jobs
type JobType string

const (
	JobTypeTranscode    JobType = "transcode"
	JobTypeNotification JobType = "notification"
	JobTypeIndexing     JobType = "indexing"
)

// TranscodeJob represents a transcoding job payload
type TranscodeJob struct {
	FileID      uint64 `json:"file_id"`
	InputPath   string `json:"input_path"`
	OutputPath  string `json:"output_path"`
	Parameters  string `json:"parameters"`
	StorageType string `json:"storage_type"` // local or s3
}
