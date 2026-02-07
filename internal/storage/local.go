package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage implements StorageService for local filesystem
type LocalStorage struct {
	basePath       string
	maxFilesPerDir int
}

// NewLocalStorage creates a new local storage service
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create base path if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &LocalStorage{
		basePath:       basePath,
		maxFilesPerDir: 65535, // Max files per data directory
	}, nil
}

// Upload uploads a file to local storage with MD5-based directory organization
func (s *LocalStorage) Upload(ctx context.Context, filename string, content io.Reader, metadata map[string]string) (string, error) {
	// Generate MD5 hash for subdirectory name (filename + timestamp)
	subdirHash := s.generateSubdirHash(filename, time.Now())
	
	// Read content and generate MD5 hash for filename
	data, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}
	
	contentHash := fmt.Sprintf("%x", md5.Sum(data))
	ext := filepath.Ext(filename)
	storedFilename := contentHash + ext
	
	// Determine which data directory to use (data1, data2, etc.)
	dataDir := s.getDataDirectory()
	
	// Create full path: basePath/dataN/subdirHash/contentHash.ext
	fullPath := filepath.Join(s.basePath, dataDir, subdirHash)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	
	filePath := filepath.Join(fullPath, storedFilename)
	
	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	
	// Return relative path: dataN/subdirHash/contentHash.ext
	relativePath := filepath.Join(dataDir, subdirHash, storedFilename)
	return relativePath, nil
}

// Download retrieves a file from local storage
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	
	return file, nil
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// Exists checks if a file exists
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// GetURL returns a local file path URL
func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	// For local storage, return the relative path
	// In a real application, this might return a URL served by the web server
	return "/storage/" + path, nil
}

// generateSubdirHash generates MD5 hash for subdirectory name
func (s *LocalStorage) generateSubdirHash(filename string, timestamp time.Time) string {
	input := filename + timestamp.Format("20060102150405")
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// getDataDirectory determines which data directory to use
func (s *LocalStorage) getDataDirectory() string {
	// Simple implementation: check existing directories and create new ones as needed
	// In production, this should track file counts and create new directories when maxFilesPerDir is reached
	
	for i := 1; i <= 100; i++ { // Check up to 100 data directories
		dir := fmt.Sprintf("data%d", i)
		dirPath := filepath.Join(s.basePath, dir)
		
		// If directory doesn't exist, create and use it
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0755); err == nil {
				return dir
			}
		}
		
		// Count files in directory
		files, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		
		// If directory has space, use it
		if len(files) < s.maxFilesPerDir {
			return dir
		}
	}
	
	// Default to data1 if all else fails
	return "data1"
}

// NormalizePath normalizes path separators for compatibility
func (s *LocalStorage) NormalizePath(path string) string {
	// Convert backslashes to forward slashes for consistency
	return strings.ReplaceAll(path, "\\", "/")
}
