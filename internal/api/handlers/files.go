package handlers

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/queue"
	"github.com/openwan/media-asset-management/internal/service"
	"github.com/openwan/media-asset-management/internal/storage"
	"github.com/openwan/media-asset-management/internal/transcoding"
)

// FileHandler handles file operations
type FileHandler struct {
	fileService    *service.FileService
	storageService storage.StorageService
	queueService   queue.QueueService
	allowedTypes   map[string][]string
	maxFileSize    int64
}

// NewFileHandler creates a new file handler
func NewFileHandler(fileService *service.FileService, storageService storage.StorageService, queueService queue.QueueService) *FileHandler {
	// Define allowed file types per category
	allowedTypes := map[string][]string{
		"video": {".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".mpg", ".mpeg"},
		"audio": {".mp3", ".wav", ".wma", ".aac", ".flac", ".ogg", ".m4a"},
		"image": {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp"},
		"rich":  {".swf", ".pdf", ".doc", ".docx", ".ppt", ".pptx"},
	}

	return &FileHandler{
		fileService:    fileService,
		storageService: storageService,
		queueService:   queueService,
		allowedTypes:   allowedTypes,
		maxFileSize:    500 * 1024 * 1024, // 500MB default
	}
}

// UploadRequest represents file upload metadata
type UploadRequest struct {
	CategoryID uint   `form:"category_id" binding:"required"`
	Title      string `form:"title" binding:"required"`
	Type       int    `form:"type"` // 1=video, 2=audio, 3=image, 4=rich
}

// Upload handles file upload
func (h *FileHandler) Upload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse multipart form
		if err := c.Request.ParseMultipartForm(h.maxFileSize); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "File too large or invalid form data",
				"error":   err.Error(),
			})
			return
		}

		// Get uploaded file
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "No file uploaded",
				"error":   err.Error(),
			})
			return
		}
		defer file.Close()

		// Parse request metadata
		var req UploadRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request parameters",
				"error":   err.Error(),
			})
			return
		}

		// Validate file size
		if header.Size > h.maxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("File size exceeds maximum allowed size of %d MB", h.maxFileSize/(1024*1024)),
			})
			return
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "File must have an extension",
			})
			return
		}

		// Determine file type if not provided
		fileType := req.Type
		if fileType == 0 {
			fileType = h.determineFileType(ext)
			if fileType == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": fmt.Sprintf("Unsupported file type: %s", ext),
				})
				return
			}
		}

		// Validate file type
		if !h.validateFileType(fileType, ext) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("File extension %s not allowed for type %d", ext, fileType),
			})
			return
		}

		// Calculate MD5 hash of file content
		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to process file",
				"error":   err.Error(),
			})
			return
		}
		md5Hash := fmt.Sprintf("%x", hash.Sum(nil))

		// Reset file pointer
		file.Seek(0, 0)

		// Generate storage path (MD5-based directory structure)
		timestamp := time.Now().Unix()
		dirHash := md5.Sum([]byte(fmt.Sprintf("%s_%d", header.Filename, timestamp)))
		dirName := fmt.Sprintf("%x", dirHash)
		fileName := fmt.Sprintf("%s%s", md5Hash, ext)
		storagePath := filepath.Join(dirName, fileName)

		// Prepare metadata for storage
		metadata := map[string]string{
			"original-filename": header.Filename,
			"content-type":      header.Header.Get("Content-Type"),
			"title":             req.Title,
		}

		// Upload file to storage
		uploadedPath, err := h.storageService.Upload(c.Request.Context(), storagePath, file, metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to upload file to storage",
				"error":   err.Error(),
			})
			return
		}

		// Get current user
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		
		// Get username string safely
		var usernameStr string
		if username != nil {
			usernameStr = username.(string)
		} else {
			usernameStr = "anonymous" // Default username if not authenticated
		}

		// Create file record
		fileRecord := &models.Files{
			CategoryID:     int(req.CategoryID),
			Type:           fileType,
			Title:          req.Title,
			Name:           md5Hash,
			Ext:            ext,
			Size:           header.Size,
			Path:           uploadedPath,
			Status:         0, // New
			Level:          1, // Default level
			UploadUsername: usernameStr,
			UploadAt:       int(time.Now().Unix()),
			Groups:         "all", // Default to all groups
		}

		if userID != nil {
			// Get user's group for access control
			// TODO: Implement group lookup
			fileRecord.Groups = "all" // Default to all groups for now
		}

		// Save file record to database
		if err := h.fileService.CreateFile(c.Request.Context(), fileRecord); err != nil {
			// Check if it's a duplicate file error
			if dupErr, ok := err.(*service.DuplicateFileError); ok {
				// Return conflict status with existing file info
				c.JSON(http.StatusConflict, gin.H{
					"success": false,
					"message": dupErr.Message,
					"code":    "DUPLICATE_FILE",
					"data": gin.H{
						"existing_file_id":    dupErr.ExistingFile.ID,
						"existing_file_title": dupErr.ExistingFile.Title,
						"existing_file_name":  dupErr.ExistingFile.Name + dupErr.ExistingFile.Ext,
						"uploaded_by":         dupErr.ExistingFile.UploadUsername,
						"uploaded_at":         dupErr.ExistingFile.UploadAt,
						"category_name":       dupErr.ExistingFile.CategoryName,
					},
				})
				return
			}
			
			// Other errors: rollback by deleting uploaded file
			h.storageService.Delete(c.Request.Context(), uploadedPath)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to save file record",
				"error":   err.Error(),
			})
			return
		}

		// Trigger transcoding for video and audio files
		if h.queueService != nil && (fileRecord.Type == 1 || fileRecord.Type == 2) {
			// Determine storage type
			storageType := "local"
			if h.storageService != nil {
				// Check if S3 storage by checking type
				storageType = "s3" // TODO: get from config
			}

			// Create transcode job
			transcodeJob := queue.TranscodeJob{
				FileID:      uint64(fileRecord.ID),
				InputPath:   uploadedPath,
				OutputPath:  strings.TrimSuffix(uploadedPath, filepath.Ext(uploadedPath)) + "-preview.flv",
				Parameters:  "-y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240",
				StorageType: storageType,
			}

			// Marshal to JSON
			jobData, err := json.Marshal(transcodeJob)
			if err == nil {
				// Create message for queue
				message := &queue.Message{
					ID:        fmt.Sprintf("transcode-%d-%d", fileRecord.ID, time.Now().Unix()),
					Body:      string(jobData),
					Timestamp: time.Now(),
					Attributes: map[string]string{
						"file_id":   strconv.FormatUint(uint64(fileRecord.ID), 10),
						"file_type": strconv.Itoa(fileRecord.Type),
					},
				}
				
				// Try to publish to queue (non-blocking)
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := h.queueService.Publish(ctx, "openwan_transcoding_jobs", message); err != nil {
						fmt.Printf("⚠ Queue unavailable for file %d, using sync transcode: %v\n", fileRecord.ID, err)
						// Queue is not available, trigger sync transcode
						h.syncTranscodeVideo(fileRecord, uploadedPath, storageType)
					} else {
						fmt.Printf("✓ Transcode job published for file %d (type %d)\n", fileRecord.ID, fileRecord.Type)
					}
				}()
			}
		} else if fileRecord.Type == 1 || fileRecord.Type == 2 {
			// No queue service, do sync transcode for video/audio
			storageType := "s3" // Assuming S3 for now
			fmt.Printf("⚠ No queue service available, using sync transcode for file %d\n", fileRecord.ID)
			go h.syncTranscodeVideo(fileRecord, uploadedPath, storageType)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "File uploaded successfully",
			"file": gin.H{
				"id":       fileRecord.ID,
				"title":    fileRecord.Title,
				"name":     fileRecord.Name,
				"ext":      fileRecord.Ext,
				"size":     fileRecord.Size,
				"type":     fileRecord.Type,
				"status":   fileRecord.Status,
				"path":     fileRecord.Path,
				"uploaded": fileRecord.UploadAt,
			},
		})
	}
}

// determineFileType determines file type based on extension
func (h *FileHandler) determineFileType(ext string) int {
	for _, videoExt := range h.allowedTypes["video"] {
		if ext == videoExt {
			return 1 // Video
		}
	}
	for _, audioExt := range h.allowedTypes["audio"] {
		if ext == audioExt {
			return 2 // Audio
		}
	}
	for _, imageExt := range h.allowedTypes["image"] {
		if ext == imageExt {
			return 3 // Image
		}
	}
	for _, richExt := range h.allowedTypes["rich"] {
		if ext == richExt {
			return 4 // Rich media
		}
	}
	return 0 // Unknown
}

// validateFileType validates if extension is allowed for file type
func (h *FileHandler) validateFileType(fileType int, ext string) bool {
	var allowedExts []string
	switch fileType {
	case 1:
		allowedExts = h.allowedTypes["video"]
	case 2:
		allowedExts = h.allowedTypes["audio"]
	case 3:
		allowedExts = h.allowedTypes["image"]
	case 4:
		allowedExts = h.allowedTypes["rich"]
	default:
		return false
	}

	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

// ListFiles returns paginated list of files
func (h *FileHandler) ListFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		// Build filter - only add filters when explicitly provided
		filter := make(map[string]interface{})
		
		// Category filter
		if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
			if categoryID, err := strconv.Atoi(categoryIDStr); err == nil && categoryID > 0 {
				filter["category_id"] = categoryID
			}
		}
		
		// File type filter
		if typeStr := c.Query("type"); typeStr != "" {
			if fileType, err := strconv.Atoi(typeStr); err == nil && fileType > 0 {
				filter["type"] = fileType
			}
		}
		
		// Status filter - only apply when parameter is explicitly provided
		if statusStr := c.Query("status"); statusStr != "" {
			if status, err := strconv.Atoi(statusStr); err == nil && status >= 0 {
				filter["status"] = status
			}
		}

		// Get files
		files, total, err := h.fileService.ListFiles(c.Request.Context(), filter, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve files",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    files,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// GetFile returns file details by ID
func (h *FileHandler) GetFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		file, err := h.fileService.GetFileByID(c.Request.Context(), uint(fileID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "File not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    file,
		})
	}
}

// DownloadFile handles file download with permission check
func (h *FileHandler) DownloadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		// Get file record
		file, err := h.fileService.GetFileByID(c.Request.Context(), uint(fileID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "File not found",
			})
			return
		}

		// Check download permission
		if !file.IsDownload {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "File download is not allowed",
			})
			return
		}

		// TODO: Check user level and group permissions

		// Get file from storage
		reader, err := h.storageService.Download(c.Request.Context(), file.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to download file",
				"error":   err.Error(),
			})
			return
		}
		defer reader.Close()

		// Get actual file size if reader is *os.File
		var fileSize int64 = int64(file.Size) // Default to database size
		if osFile, ok := reader.(*os.File); ok {
			if stat, err := osFile.Stat(); err == nil {
				fileSize = stat.Size() // Use actual file size
			}
		}

	// Set headers for download
	// Use title as filename if available, otherwise use MD5 name
	var filename string
	if file.Title != "" {
		// Sanitize title to remove invalid filename characters
		sanitizedTitle := strings.Map(func(r rune) rune {
			// Allow alphanumeric, Chinese, spaces, hyphens, underscores, periods
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
				r >= 0x4E00 && r <= 0x9FFF || // Chinese characters
				r == ' ' || r == '-' || r == '_' || r == '.' {
				return r
			}
			return '_'
		}, file.Title)
		filename = fmt.Sprintf("%s%s", sanitizedTitle, file.Ext)
	} else {
		filename = fmt.Sprintf("%s%s", file.Name, file.Ext)
	}
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Type", getContentType(file.Ext))
		c.Header("Content-Length", fmt.Sprintf("%d", fileSize))

		// Stream file to client
		io.Copy(c.Writer, reader)
	}
}

// PreviewFile serves the preview version of a file (for video/audio, serves FLV preview or falls back to original)
func (h *FileHandler) PreviewFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		// Get file record
		file, err := h.fileService.GetFileByID(c.Request.Context(), uint(fileID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "File not found",
			})
			return
		}

		// Determine preview file path and content type
		var contentType string
		var reader io.ReadCloser
		
		// For video and audio files, try preview first, then fall back to original
		if file.Type == models.FileTypeVideo || file.Type == models.FileTypeAudio {
			// Try preview file: replace extension with -preview.flv
			// Use string manipulation to handle S3 paths correctly
			ext := filepath.Ext(file.Path)
			previewPath := strings.TrimSuffix(file.Path, ext) + "-preview.flv"
			
			reader, err = h.storageService.Download(c.Request.Context(), previewPath)
			if err != nil {
				// Preview not available, fall back to original file
				reader, err = h.storageService.Download(c.Request.Context(), file.Path)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{
						"success": false,
						"message": "File not available",
						"error":   err.Error(),
					})
					return
				}
				// Set content type based on original file extension
				ext := strings.ToLower(file.Ext)
				switch ext {
				case ".mp4":
					contentType = "video/mp4"
				case ".avi":
					contentType = "video/x-msvideo"
				case ".mov":
					contentType = "video/quicktime"
				case ".mkv":
					contentType = "video/x-matroska"
				case ".flv":
					contentType = "video/x-flv"
				case ".mp3":
					contentType = "audio/mpeg"
				case ".wav":
					contentType = "audio/wav"
				case ".aac":
					contentType = "audio/aac"
				default:
					contentType = "video/mp4"
				}
			} else {
				// Preview file exists
				contentType = "video/x-flv"
			}
		} else if file.Type == models.FileTypeImage {
			// For images, use the original file
			reader, err = h.storageService.Download(c.Request.Context(), file.Path)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Image file not available",
					"error":   err.Error(),
				})
				return
			}
			// Determine image content type based on extension
			ext := strings.ToLower(file.Ext)
			switch ext {
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			case ".gif":
				contentType = "image/gif"
			case ".bmp":
				contentType = "image/bmp"
			case ".webp":
				contentType = "image/webp"
			default:
				contentType = "image/jpeg"
			}
		} else {
			// Other file types don't support preview
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Preview not supported for this file type",
			})
			return
		}
		defer reader.Close()

		// Set headers for streaming
		c.Header("Content-Type", contentType)
		c.Header("Accept-Ranges", "bytes")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("X-Content-Type-Options", "nosniff")

		// Stream file to client
		io.Copy(c.Writer, reader)
	}
}

// UpdateFile updates file metadata
func (h *FileHandler) UpdateFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.fileService.UpdateFile(c.Request.Context(), uint(fileID), updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update file",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "File updated successfully",
		})
	}
}

// DeleteFile deletes a file (soft delete by setting status to 4)
func (h *FileHandler) DeleteFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		// Get username from session (if authenticated)
		username := ""
		if user, exists := c.Get("user"); exists {
			if userMap, ok := user.(map[string]interface{}); ok {
				if uname, ok := userMap["username"].(string); ok {
					username = uname
				}
			}
		}

		// Hard delete - permanently remove from database
		if err := h.fileService.DeleteFile(c.Request.Context(), uint64(fileID), username); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete file",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "File deleted successfully",
		})
	}
}


// getContentType returns the appropriate MIME type based on file extension
func getContentType(ext string) string {
	ext = strings.ToLower(ext)
	
	// Video formats
	videoTypes := map[string]string{
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".mkv":  "video/x-matroska",
		".mpg":  "video/mpeg",
		".mpeg": "video/mpeg",
		".webm": "video/webm",
	}
	
	// Audio formats
	audioTypes := map[string]string{
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".wma":  "audio/x-ms-wma",
		".aac":  "audio/aac",
		".flac": "audio/flac",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
	}
	
	// Image formats
	imageTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".tiff": "image/tiff",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
	}
	
	// Document formats
	docTypes := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".swf":  "application/x-shockwave-flash",
	}
	
	// Check each type map
	if contentType, ok := videoTypes[ext]; ok {
		return contentType
	}
	if contentType, ok := audioTypes[ext]; ok {
		return contentType
	}
	if contentType, ok := imageTypes[ext]; ok {
		return contentType
	}
	if contentType, ok := docTypes[ext]; ok {
		return contentType
	}
	
	// Default to octet-stream for unknown types
	return "application/octet-stream"
}

// GetStats returns file statistics
func (h *FileHandler) GetStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get statistics from service
		stats, err := h.fileService.GetStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get file statistics",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    stats,
		})
	}
}

// GetRecentFiles handles recent files listing
func (h *FileHandler) GetRecentFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get limit from query parameter
		limit := 10 // default
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		// Get recent files from service
		files, err := h.fileService.GetRecentFiles(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get recent files",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    files,
			"total":   len(files),
		})
	}
}

// syncTranscodeVideo performs synchronous video transcoding
// This is a fallback when RabbitMQ queue is not available
func (h *FileHandler) syncTranscodeVideo(fileRecord *models.Files, inputPath, storageType string) {
	fmt.Printf("🎬 Starting sync transcode for file %d (%s)\n", fileRecord.ID, inputPath)
	
	// Create FFmpeg wrapper
	ffmpegWrapper := transcoding.NewFFmpegWrapper("/usr/local/bin/ffmpeg", "")
	
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "-preview.flv"
	parameters := "-y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240"
	
	var inputFile, outputFile string
	
	if storageType == "s3" {
		// Download from S3 to temp file
		tempDir := "/tmp/openwan-transcode"
		os.MkdirAll(tempDir, 0755)
		
		inputFile = filepath.Join(tempDir, fmt.Sprintf("input-%d%s", fileRecord.ID, filepath.Ext(inputPath)))
		outputFile = filepath.Join(tempDir, fmt.Sprintf("output-%d.flv", fileRecord.ID))
		
		fmt.Printf("  ⬇  Downloading from S3: %s -> %s\n", inputPath, inputFile)
		
		// Download from S3
		ctx := context.Background()
		reader, err := h.storageService.Download(ctx, inputPath)
		if err != nil {
			fmt.Printf("  ❌ Failed to download from S3: %v\n", err)
			return
		}
		defer reader.Close()
		
		// Write to temp file
		f, err := os.Create(inputFile)
		if err != nil {
			fmt.Printf("  ❌ Failed to create temp file: %v\n", err)
			return
		}
		defer f.Close()
		defer os.Remove(inputFile) // Clean up
		
		if _, err := io.Copy(f, reader); err != nil {
			fmt.Printf("  ❌ Failed to write temp file: %v\n", err)
			return
		}
		f.Close()
		
		fmt.Printf("  ✓ Downloaded %s\n", inputFile)
	} else {
		inputFile = inputPath
		outputFile = outputPath
	}
	
	// Transcode
	fmt.Printf("  🎥 Transcoding: %s -> %s\n", inputFile, outputFile)
	fmt.Printf("  📝 Parameters: %s\n", parameters)
	
	ctx := context.Background()
	opts := transcoding.TranscodeOptions{
		InputPath:    inputFile,
		OutputPath:   outputFile,
		CustomParams: parameters,
	}
	
	if err := ffmpegWrapper.Transcode(ctx, opts); err != nil {
		fmt.Printf("  ❌ Transcode failed: %v\n", err)
		return
	}
	
	fmt.Printf("  ✓ Transcode completed: %s\n", outputFile)
	
	// Upload output to S3 if needed
	if storageType == "s3" {
		defer os.Remove(outputFile) // Clean up
		
		fmt.Printf("  ⬆  Uploading to S3: %s -> %s\n", outputFile, outputPath)
		
		f, err := os.Open(outputFile)
		if err != nil {
			fmt.Printf("  ❌ Failed to open output file: %v\n", err)
			return
		}
		defer f.Close()
		
		ctx := context.Background()
		metadata := map[string]string{
			"Content-Type":  "video/x-flv",
			"original-file": strconv.FormatUint(uint64(fileRecord.ID), 10),
			"transcode-date": time.Now().Format(time.RFC3339),
		}
		
		if _, err := h.storageService.Upload(ctx, outputPath, f, metadata); err != nil {
			fmt.Printf("  ❌ Failed to upload to S3: %v\n", err)
			return
		}
		
		fmt.Printf("  ✓ Uploaded preview to S3: %s\n", outputPath)
	}
	
	fmt.Printf("✅ Transcode completed for file %d\n", fileRecord.ID)
}
