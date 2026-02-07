package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/repository"
)

// DuplicateFileError represents a duplicate file error with existing file information
type DuplicateFileError struct {
	Message      string
	ExistingFile *models.Files
}

func (e *DuplicateFileError) Error() string {
	return e.Message
}

// FilesService handles file-related business logic
type FilesService struct {
	repo repository.Repository
}

// FileService is an alias for FilesService for handler compatibility
type FileService = FilesService

// NewFilesService creates a new files service
func NewFilesService(repo repository.Repository) *FilesService {
	return &FilesService{repo: repo}
}

// NewFileService creates a new file service (alias)
func NewFileService(repo repository.Repository) *FileService {
	return NewFilesService(repo)
}

// CreateFile creates a new file record
func (s *FilesService) CreateFile(ctx context.Context, file *models.Files) error {
	// Validate file type based on extension
	if err := s.ValidateFileType(file.Ext, file.Type); err != nil {
		return err
	}

	// Check for MD5 collision - return existing file info if duplicate
	existing, err := s.repo.Files().FindByMD5(ctx, file.Name)
	if err != nil {
		return fmt.Errorf("failed to check MD5: %w", err)
	}
	if existing != nil {
		return &DuplicateFileError{
			Message:      "文件已存在，这是重复文件",
			ExistingFile: existing,
		}
	}

	return s.repo.Files().Create(ctx, file)
}

// GetFile retrieves a file by ID with access control
func (s *FilesService) GetFile(ctx context.Context, fileID uint64, userID int) (*models.Files, error) {
	file, err := s.repo.Files().FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// Check access permission
	canAccess, err := s.repo.ACL().CanAccessFile(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, fmt.Errorf("access denied to file %d", fileID)
	}

	return file, nil
}

// ListFiles retrieves files with filters and pagination
func (s *FilesService) ListFiles(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Files, int64, error) {
	// Calculate offset from page number
	offset := (page - 1) * pageSize
	
	// For now, skip user access level filtering if not provided
	// This should be added by the caller if needed
	return s.repo.Files().FindAll(ctx, filters, pageSize, offset)
}

// GetFileByID retrieves a file by ID
func (s *FilesService) GetFileByID(ctx context.Context, fileID uint) (*models.Files, error) {
	return s.repo.Files().FindByID(ctx, uint64(fileID))
}

// UpdateFile updates file information
func (s *FilesService) UpdateFile(ctx context.Context, fileID uint, updates map[string]interface{}) error {
	// Get existing file
	file, err := s.repo.Files().FindByID(ctx, uint64(fileID))
	if err != nil {
		return err
	}
	
	// Apply updates
	// This is a simplified version - in production you'd use reflection or a proper update method
	if title, ok := updates["title"].(string); ok {
		file.Title = title
	}
	if status, ok := updates["status"].(int); ok {
		file.Status = status
	}
	if level, ok := updates["level"].(int); ok {
		file.Level = level
	}
	if groups, ok := updates["groups"].(string); ok {
		file.Groups = groups
	}
	if catalogInfo, ok := updates["catalog_info"].(string); ok {
		file.CatalogInfo = catalogInfo
	}
	if isDownload, ok := updates["is_download"].(bool); ok {
		file.IsDownload = isDownload
	}
	
	return s.repo.Files().Update(ctx, file)
}

// UpdateFileMetadata updates file metadata
func (s *FilesService) UpdateFileMetadata(ctx context.Context, file *models.Files) error {
	return s.repo.Files().Update(ctx, file)
}

// SubmitForReview changes file status from new to pending
func (s *FilesService) SubmitForReview(ctx context.Context, fileID uint64, username string) error {
	return s.repo.Files().UpdateStatus(ctx, fileID, models.FileStatusPending, username)
}

// PublishFile changes file status to published
func (s *FilesService) PublishFile(ctx context.Context, fileID uint64, username string) error {
	return s.repo.Files().UpdateStatus(ctx, fileID, models.FileStatusPublished, username)
}

// RejectFile changes file status to rejected
func (s *FilesService) RejectFile(ctx context.Context, fileID uint64, username string) error {
	return s.repo.Files().UpdateStatus(ctx, fileID, models.FileStatusRejected, username)
}

// DeleteFile soft deletes a file
func (s *FilesService) DeleteFile(ctx context.Context, fileID uint64, username string) error {
	// Hard delete - permanently remove from database
	return s.repo.Files().Delete(ctx, fileID)
}

// ValidateFileType validates file extension against file type
func (s *FilesService) ValidateFileType(ext string, fileType int) error {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	
	videoExts := []string{"mp4", "avi", "flv", "wmv", "mov", "mkv", "mpg", "mpeg", "asf"}
	audioExts := []string{"mp3", "wav", "wma", "aac", "flac", "ogg", "m4a"}
	imageExts := []string{"jpg", "jpeg", "png", "gif", "bmp", "tif", "tiff", "svg", "webp"}
	richExts := []string{"swf", "html", "htm", "zip", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf", "txt", "lrc"}

	var validExts []string
	switch fileType {
	case models.FileTypeVideo:
		validExts = videoExts
	case models.FileTypeAudio:
		validExts = audioExts
	case models.FileTypeImage:
		validExts = imageExts
	case models.FileTypeRichMedia:
		validExts = richExts
	default:
		return fmt.Errorf("invalid file type: %d", fileType)
	}

	for _, validExt := range validExts {
		if ext == validExt {
			return nil
		}
	}

	return fmt.Errorf("extension %s is not valid for file type %d", ext, fileType)
}

// DetectFileType detects file type from extension
func (s *FilesService) DetectFileType(filename string) int {
	ext := strings.ToLower(filepath.Ext(filename))
	ext = strings.TrimPrefix(ext, ".")

	videoExts := map[string]bool{"mp4": true, "avi": true, "flv": true, "wmv": true, "mov": true, "mkv": true}
	audioExts := map[string]bool{"mp3": true, "wav": true, "wma": true, "aac": true, "flac": true}
	imageExts := map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true, "bmp": true}

	if videoExts[ext] {
		return models.FileTypeVideo
	}
	if audioExts[ext] {
		return models.FileTypeAudio
	}
	if imageExts[ext] {
		return models.FileTypeImage
	}
	return models.FileTypeRichMedia
}

// FileStats represents file statistics
type FileStats struct {
	Total     int64 `json:"total"`
	Video     int64 `json:"video"`
	Audio     int64 `json:"audio"`
	Image     int64 `json:"image"`
	RichMedia int64 `json:"rich_media"`
	NewCount  int64 `json:"new_count"`
	Published int64 `json:"published"`
	Pending   int64 `json:"pending"`
}

// GetStats returns file statistics
func (s *FilesService) GetStats() (*FileStats, error) {
	stats := &FileStats{}
	
	// Use Files() to get FilesRepository, which has access to db
	// We need to use the repository pattern correctly
	// Let's create a custom query in the repository
	
	// For now, use a simple approach - count through list operations
	// This is not optimal but works without modifying repository interface
	ctx := context.Background()
	
	// Get total count by fetching all with high limit
	allFiles, total, err := s.repo.Files().FindAll(ctx, map[string]interface{}{}, 100000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to count total files: %w", err)
	}
	stats.Total = total
	
	// Count by type and status from the fetched files
	for _, file := range allFiles {
		switch file.Type {
		case models.FileTypeVideo:
			stats.Video++
		case models.FileTypeAudio:
			stats.Audio++
		case models.FileTypeImage:
			stats.Image++
		case models.FileTypeRichMedia:
			stats.RichMedia++
		}
		
		switch file.Status {
		case models.FileStatusNew:
			stats.NewCount++
		case models.FileStatusPublished:
			stats.Published++
		case models.FileStatusPending:
			stats.Pending++
		}
	}
	
	return stats, nil
}

// GetRecentFiles retrieves recent files ordered by upload date
func (s *FilesService) GetRecentFiles(ctx context.Context, limit int) ([]*models.Files, error) {
	// Default limit to 10 if not specified or invalid
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	
	// Use FindAll with ordering by upload_at descending
	filters := map[string]interface{}{
		"order": "upload_at DESC",
	}
	
	files, _, err := s.repo.Files().FindAll(ctx, filters, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent files: %w", err)
	}
	
	return files, nil
}
