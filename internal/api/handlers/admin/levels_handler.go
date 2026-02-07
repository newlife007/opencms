package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/service"
)

// LevelsHandler handles levels management endpoints
type LevelsHandler struct {
	service *service.LevelsService
}

// NewLevelsHandler creates a new levels handler
func NewLevelsHandler(service *service.LevelsService) *LevelsHandler {
	return &LevelsHandler{
		service: service,
	}
}

// ListLevels returns all levels
func (h *LevelsHandler) ListLevels(c *gin.Context) {
	levels, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve levels",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": levels,
		"total": len(levels),
	})
}

// GetLevel returns a single level by ID
func (h *LevelsHandler) GetLevel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid level ID",
		})
		return
	}

	level, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Level not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": level,
	})
}

// CreateLevel creates a new level
func (h *LevelsHandler) CreateLevel(c *gin.Context) {
	var level models.Levels
	if err := c.ShouldBindJSON(&level); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error": err.Error(),
		})
		return
	}

	if err := h.service.Create(c.Request.Context(), &level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create level",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Level created successfully",
		"data": level,
	})
}

// UpdateLevel updates an existing level
func (h *LevelsHandler) UpdateLevel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid level ID",
		})
		return
	}

	var level models.Levels
	if err := c.ShouldBindJSON(&level); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error": err.Error(),
		})
		return
	}

	level.ID = id
	if err := h.service.Update(c.Request.Context(), &level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update level",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Level updated successfully",
		"data": level,
	})
}

// DeleteLevel deletes a level
func (h *LevelsHandler) DeleteLevel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid level ID",
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete level",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Level deleted successfully",
	})
}
