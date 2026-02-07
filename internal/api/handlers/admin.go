package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/service"
)

// AdminHandler handles admin operations
type AdminHandler struct {
	aclService *service.ACLService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(aclService *service.ACLService) *AdminHandler {
	return &AdminHandler{
		aclService: aclService,
	}
}

// ListUsers returns paginated list of users
func (h *AdminHandler) ListUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		offset := (page - 1) * pageSize
		users, total, err := h.aclService.ListUsers(c.Request.Context(), pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve users",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    users,
			"pagination": gin.H{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// GetUser returns user details by ID
func (h *AdminHandler) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid user ID",
			})
			return
		}

		user, err := h.aclService.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		// Get user permissions
		permissions, _ := h.aclService.GetUserPermissions(c.Request.Context(), user.ID)
		
		permList := make([]string, len(permissions))
		for i, p := range permissions {
			// Build permission name from namespace/controller/action
			permList[i] = p.Namespace + "." + p.Controller + "." + p.Action
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"user":        user,
				"permissions": permList,
			},
		})
	}
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		type CreateUserRequest struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required,min=6"`
			Email    string `json:"email" binding:"required,email"`
			GroupID  uint   `json:"group_id" binding:"required"`
			LevelID  uint   `json:"level_id" binding:"required"`
			Enabled  bool   `json:"enabled"`
		}

		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		user := &models.Users{
			Username: req.Username,
			Email:    &req.Email,
			GroupID:  int(req.GroupID),
			LevelID:  int(req.LevelID),
			Enabled:  req.Enabled,
		}

		if err := h.aclService.CreateUser(c.Request.Context(), user, req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create user",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "User created successfully",
			"data":    user,
		})
	}
}

// UpdateUser updates user information
func (h *AdminHandler) UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid user ID",
			})
			return
		}

		user, err := h.aclService.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		type UpdateUserRequest struct {
			Email   string `json:"email"`
			GroupID uint   `json:"group_id"`
			LevelID uint   `json:"level_id"`
			Enabled *bool  `json:"enabled"`
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		// Update fields
		if req.Email != "" {
			user.Email = &req.Email
		}
		if req.GroupID > 0 {
			user.GroupID = int(req.GroupID)
		}
		if req.LevelID > 0 {
			user.LevelID = int(req.LevelID)
		}
		if req.Enabled != nil {
			user.Enabled = *req.Enabled
		}

		if err := h.aclService.UpdateUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update user",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User updated successfully",
			"data":    user,
		})
	}
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid user ID",
			})
			return
		}

		if err := h.aclService.DeleteUser(c.Request.Context(), userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete user",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User deleted successfully",
		})
	}
}

// ResetPassword resets a user's password
func (h *AdminHandler) ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid user ID",
			})
			return
		}

		type ResetPasswordRequest struct {
			NewPassword string `json:"new_password" binding:"required,min=6"`
		}

		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.aclService.ResetPassword(c.Request.Context(), userID, req.NewPassword); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to reset password",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password reset successfully",
		})
	}
}
