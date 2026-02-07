package models

import "time"

// TranscodeJob represents a transcoding job status in database
type TranscodeJob struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FileID          uint64    `gorm:"column:file_id;not null;index" json:"file_id"`
	InputPath       string    `gorm:"column:input_path;type:varchar(255);not null" json:"input_path"`
	OutputPath      string    `gorm:"column:output_path;type:varchar(255);not null" json:"output_path"`
	Status          string    `gorm:"column:status;type:varchar(32);not null;index" json:"status"` // pending, processing, completed, failed
	Progress        float64   `gorm:"column:progress;not null;default:0" json:"progress"`
	WorkerID        string    `gorm:"column:worker_id;type:varchar(64)" json:"worker_id,omitempty"`
	ErrorMessage    string    `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	RetryCount      int       `gorm:"column:retry_count;not null;default:0" json:"retry_count"`
	StartedAt       *time.Time `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt     *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for TranscodeJob
func (TranscodeJob) TableName() string {
	return "ow_transcode_jobs"
}

// Job status constants
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)
