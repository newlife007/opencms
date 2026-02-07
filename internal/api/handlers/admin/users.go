package admin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/service"
)

// UsersHandler handles user management
type UsersHandler struct {
	usersService *service.UsersService
}

// NewUsersHandler creates a new users handler
func NewUsersHandler(usersService *service.UsersService) *UsersHandler {
	return &UsersHandler{usersService: usersService}
}

// ListUsers handles GET/POST /api/v1/admin/users
// Supports both GET (query params) and POST (JSON body) for flexibility
func (h *UsersHandler) ListUsers(c *gin.Context) {
	var req service.UserListRequest

	// Support both GET query params and POST JSON body
	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	} else {
		// Parse query parameters
		req.Username = c.Query("username")
		req.Nickname = c.Query("nickname")
		req.Email = c.Query("email")
		
		// Parse group_ids
		if groupIDsStr := c.Query("group_ids"); groupIDsStr != "" {
			// Parse comma-separated IDs
			// Note: In production, you might want to use a library for this
			// For now, we'll keep it simple
		}
		
		// Parse level_ids
		if levelIDsStr := c.Query("level_ids"); levelIDsStr != "" {
			// Parse comma-separated IDs
		}
		
		// Parse enabled
		if enabledStr := c.Query("enabled"); enabledStr != "" {
			enabled := enabledStr == "1" || enabledStr == "true"
			req.Enabled = &enabled
		}
		
		// Parse pagination
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
		req.Page = page
		req.PageSize = pageSize
	}

	users, total, err := h.usersService.ListUsers(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve users",
			"error":   err.Error(),
		})
		return
	}

	// Calculate pagination info
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"pagination": gin.H{
			"total":       total,
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
		},
	})
}

// GetUser handles GET /api/v1/admin/users/:id
func (h *UsersHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	user, err := h.usersService.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "success",
		"data": user,
	})
}

// CreateUser handles POST /api/v1/admin/users
func (h *UsersHandler) CreateUser(c *gin.Context) {
	// Log raw request body for debugging
	bodyBytes, _ := c.GetRawData()
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	var req service.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Return detailed error with request body for debugging
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
			"debug_request_body": string(bodyBytes),
			"debug_error": fmt.Sprintf("%v", err),
		})
		return
	}

	user, err := h.usersService.CreateUser(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		
		// Handle specific errors
		switch err {
		case service.ErrDuplicateUsername:
			statusCode = http.StatusConflict
			msg = "Username already exists"
		case service.ErrWeakPassword:
			statusCode = http.StatusBadRequest
			msg = "Password must be 3-32 characters"
		}
		
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": msg,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully",
		"data": user,
	})
}

// UpdateUser handles PUT /api/v1/admin/users/:id
func (h *UsersHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	var req service.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	req.ID = id

	user, err := h.usersService.UpdateUser(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		
		switch err {
		case service.ErrUserNotFound:
			statusCode = http.StatusNotFound
			msg = "User not found"
		case service.ErrWeakPassword:
			statusCode = http.StatusBadRequest
			msg = "Password must be 3-32 characters"
		}
		
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User updated successfully",
		"data": user,
	})
}

// DeleteUser handles DELETE /api/v1/admin/users/:id
func (h *UsersHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	err = h.usersService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		
		if err == service.ErrUserNotFound {
			statusCode = http.StatusNotFound
			msg = "User not found"
		}
		
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}

// BatchDeleteUsers handles POST /api/v1/admin/users/batch-delete
func (h *UsersHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No user IDs provided",
		})
		return
	}

	err := h.usersService.BatchDeleteUsers(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Users deleted successfully",
		"data": gin.H{
			"deleted_count": len(req.IDs),
		},
	})
}



// ResetUserPassword resets a user's password (admin only)
func (h *UsersHandler) ResetUserPassword(c *gin.Context) {
	// Get user ID from path parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Reset password through service
	if err := h.usersService.ResetPassword(c.Request.Context(), uint(userID), req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to reset password: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully",
	})
}

// UpdateUserStatus updates a user's enabled status
func (h *UsersHandler) UpdateUserStatus(c *gin.Context) {
	// Get user ID from path parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req struct {
		Status *int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Validate status field is present
	if req.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: status field is required",
		})
		return
	}

	// Validate status value
	if *req.Status != 0 && *req.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: status must be 0 (disabled) or 1 (enabled)",
		})
		return
	}

	// Update status through service
	enabled := *req.Status == 1
	if err := h.usersService.UpdateUserStatus(c.Request.Context(), uint(userID), enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update user status: " + err.Error(),
		})
		return
	}

	statusText := "disabled"
	if enabled {
		statusText = "enabled"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User " + statusText + " successfully",
	})
}

// GetUserPermissions gets a user's effective permissions
func (h *UsersHandler) GetUserPermissions(c *gin.Context) {
	// Get user ID from path parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Get user permissions through service
	permissions, err := h.usersService.GetUserPermissions(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user permissions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Success",
		"data": gin.H{
			"user_id":     userID,
			"permissions": permissions,
		},
	})
}

