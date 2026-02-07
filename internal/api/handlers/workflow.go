package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/service"
)

// WorkflowHandler handles file workflow operations
type WorkflowHandler struct {
	fileService *service.FileService
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(fileService *service.FileService) *WorkflowHandler {
	return &WorkflowHandler{
		fileService: fileService,
	}
}

// SubmitForReview submits a file for review (status: 0 -> 1)
func (h *WorkflowHandler) SubmitForReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		username, _ := c.Get("username")
		usernameStr := ""
		if username != nil {
			usernameStr = username.(string)
		}

		if err := h.fileService.SubmitForReview(c.Request.Context(), fileID, usernameStr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to submit file for review",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "File submitted for review successfully",
		})
	}
}

// PublishFile publishes a file (status: 1 -> 2)
func (h *WorkflowHandler) PublishFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		username, _ := c.Get("username")
		usernameStr := ""
		if username != nil {
			usernameStr = username.(string)
		}

		if err := h.fileService.PublishFile(c.Request.Context(), fileID, usernameStr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to publish file",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "File published successfully",
		})
	}
}

// RejectFile rejects a file (status: 1 -> 3)
func (h *WorkflowHandler) RejectFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		type RejectRequest struct {
			Reason string `json:"reason"`
		}

		var req RejectRequest
		c.ShouldBindJSON(&req)

		username, _ := c.Get("username")
		usernameStr := ""
		if username != nil {
			usernameStr = username.(string)
		}

		if err := h.fileService.RejectFile(c.Request.Context(), fileID, usernameStr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to reject file",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "File rejected successfully",
		})
	}
}

// UpdateFileStatus updates file status with validation
func (h *WorkflowHandler) UpdateFileStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid file ID",
			})
			return
		}

		type StatusUpdateRequest struct {
			Status int    `json:"status" binding:"required,min=0,max=4"`
			Reason string `json:"reason"`
		}

		var req StatusUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		username, _ := c.Get("username")
		usernameStr := ""
		if username != nil {
			usernameStr = username.(string)
		}

		// Get current file
		file, err := h.fileService.GetFileByID(c.Request.Context(), uint(fileID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "File not found",
			})
			return
		}

		// Validate status transition
		if !h.isValidStatusTransition(file.Status, req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid status transition",
				"from":    file.Status,
				"to":      req.Status,
			})
			return
		}

		// Update status based on target
		var updateErr error
		switch req.Status {
		case 1: // Pending
			updateErr = h.fileService.SubmitForReview(c.Request.Context(), fileID, usernameStr)
		case 2: // Published
			updateErr = h.fileService.PublishFile(c.Request.Context(), fileID, usernameStr)
		case 3: // Rejected
			updateErr = h.fileService.RejectFile(c.Request.Context(), fileID, usernameStr)
		case 4: // Deleted
			updateErr = h.fileService.DeleteFile(c.Request.Context(), fileID, usernameStr)
		}

		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update file status",
				"error":   updateErr.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "File status updated successfully",
		})
	}
}

// isValidStatusTransition validates if status transition is allowed
func (h *WorkflowHandler) isValidStatusTransition(from, to int) bool {
	// Status: 0=new, 1=pending, 2=published, 3=rejected, 4=deleted
	validTransitions := map[int][]int{
		0: {1, 4},       // New -> Pending, Deleted
		1: {2, 3, 4},    // Pending -> Published, Rejected, Deleted
		2: {1, 4},       // Published -> Pending (for re-review), Deleted
		3: {1, 4},       // Rejected -> Pending (resubmit), Deleted
		4: {1},          // Deleted -> Pending (restore)
	}

	allowedTargets, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTargets {
		if allowed == to {
			return true
		}
	}

	return false
}

// GetWorkflowStats returns workflow statistics
func (h *WorkflowHandler) GetWorkflowStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement actual statistics query
		// This is a placeholder implementation
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"stats": gin.H{
				"new":       0,
				"pending":   0,
				"published": 0,
				"rejected":  0,
				"deleted":   0,
				"total":     0,
			},
		})
	}
}
