package models

import (
	"time"
)

// Files represents the ow_files table for media file management
type Files struct {
	ID             uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CategoryID     int    `gorm:"column:category_id;not null;index" json:"category_id"`
	CategoryName   string `gorm:"column:category_name;type:varchar(64);not null" json:"category_name"`
	Type           int    `gorm:"column:type;not null;default:1;index" json:"type"` // 1:video 2:audio 3:image 4:rich_media
	Title          string `gorm:"column:title;type:varchar(255);not null;index" json:"title"`
	Name           string `gorm:"column:name;type:varchar(255);not null" json:"name"` // MD5 filename
	Ext            string `gorm:"column:ext;type:varchar(16);not null" json:"ext"`
	Size           int64  `gorm:"column:size;not null;default:0" json:"size"`
	Path           string `gorm:"column:path;type:varchar(255);not null" json:"path"`
	Status         int    `gorm:"column:status;not null;index" json:"status"` // 0:new 1:pending 2:published 3:rejected 4:deleted
	Level          int    `gorm:"column:level;not null;default:1;index" json:"level"`
	Groups         string `gorm:"column:groups;type:varchar(255);not null;default:'all'" json:"groups"` // comma-separated group IDs or 'all'
	IsDownload     bool   `gorm:"column:is_download;not null;default:true" json:"is_download"`
	CatalogInfo    string `gorm:"column:catalog_info;type:text;not null" json:"catalog_info"` // JSON metadata
	UploadUsername string `gorm:"column:upload_username;type:varchar(64);not null" json:"upload_username"`
	UploadAt       int    `gorm:"column:upload_at;not null" json:"upload_at"` // Unix timestamp
	CatalogUsername *string `gorm:"column:catalog_username;type:varchar(64)" json:"catalog_username,omitempty"`
	CatalogAt      *int    `gorm:"column:catalog_at" json:"catalog_at,omitempty"` // Unix timestamp
	PutoutUsername *string `gorm:"column:putout_username;type:varchar(64)" json:"putout_username,omitempty"`
	PutoutAt       *int    `gorm:"column:putout_at" json:"putout_at,omitempty"` // Unix timestamp
}

// TableName specifies the table name for the Files model
func (Files) TableName() string {
	return "ow_files"
}

// BeforeCreate hook to set upload_at if not set
func (f *Files) BeforeCreate() error {
	if f.UploadAt == 0 {
		f.UploadAt = int(time.Now().Unix())
	}
	return nil
}

// FileType constants
const (
	FileTypeVideo     = 1
	FileTypeAudio     = 2
	FileTypeImage     = 3
	FileTypeRichMedia = 4
)

// FileStatus constants
const (
	FileStatusNew       = 0
	FileStatusPending   = 1
	FileStatusPublished = 2
	FileStatusRejected  = 3
	FileStatusDeleted   = 4
)
