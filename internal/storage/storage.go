package storage

import (
	"context"
	"io"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	// Upload uploads a file and returns the storage path
	Upload(ctx context.Context, filename string, content io.Reader, metadata map[string]string) (string, error)
	
	// Download retrieves a file by path
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	
	// Delete removes a file
	Delete(ctx context.Context, path string) error
	
	// Exists checks if a file exists
	Exists(ctx context.Context, path string) (bool, error)
	
	// GetURL returns a URL for accessing the file
	GetURL(ctx context.Context, path string) (string, error)
}

// Metadata keys
const (
	MetadataContentType    = "content-type"
	MetadataOriginalName   = "original-name"
	MetadataUploadUsername = "upload-username"
	MetadataCategoryID     = "category-id"
	MetadataFileType       = "file-type"
)
