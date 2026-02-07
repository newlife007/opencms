package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/service"
)

// CategoryHandler handles category operations
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// ListCategories returns hierarchical list of categories
func (h *CategoryHandler) ListCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all categories
		categories, err := h.categoryService.GetAllCategories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve categories",
				"error":   err.Error(),
			})
			return
		}

		// Build hierarchical tree
		tree := h.buildCategoryTree(categories, 0)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    tree,
		})
	}
}

// GetCategoryTree returns hierarchical tree structure of categories
// This is an alias for ListCategories for frontend compatibility
func (h *CategoryHandler) GetCategoryTree() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all categories
		categories, err := h.categoryService.GetAllCategories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve category tree",
				"error":   err.Error(),
			})
			return
		}

		// Build hierarchical tree
		tree := h.buildCategoryTree(categories, 0)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    tree,
		})
	}
}


// GetCategory returns category details by ID
func (h *CategoryHandler) GetCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid category ID",
			})
			return
		}

		category, err := h.categoryService.GetCategoryByID(c.Request.Context(), uint(categoryID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Category not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    category,
		})
	}
}

// CreateCategory creates a new category
func (h *CategoryHandler) CreateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var category models.Category
		if err := c.ShouldBindJSON(&category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.categoryService.CreateCategory(c.Request.Context(), &category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create category",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Category created successfully",
			"data":    category,
		})
	}
}

// UpdateCategory updates category details
func (h *CategoryHandler) UpdateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid category ID",
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

		if err := h.categoryService.UpdateCategory(c.Request.Context(), uint(categoryID), updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update category",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Category updated successfully",
		})
	}
}

// DeleteCategory deletes a category
func (h *CategoryHandler) DeleteCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid category ID",
			})
			return
		}

		if err := h.categoryService.DeleteCategory(c.Request.Context(), uint(categoryID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete category",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Category deleted successfully",
		})
	}
}

// CategoryNode represents a node in category tree
type CategoryNode struct {
	ID          int             `json:"id"`
	ParentID    int             `json:"parent_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Path        string          `json:"path"`
	Weight      int             `json:"weight"`
	Enabled     bool            `json:"enabled"`
	Children    []*CategoryNode `json:"children,omitempty"`
}

// buildCategoryTree builds hierarchical tree structure
func (h *CategoryHandler) buildCategoryTree(categories []models.Category, parentID int) []*CategoryNode {
	var nodes []*CategoryNode

	for _, cat := range categories {
		if cat.ParentID == parentID {
			node := &CategoryNode{
				ID:          cat.ID,
				ParentID:    cat.ParentID,
				Name:        cat.Name,
				Description: cat.Description,
				Path:        cat.Path,
				Weight:      cat.Weight,
				Enabled:     cat.Enabled,
			}
			node.Children = h.buildCategoryTree(categories, cat.ID)
			nodes = append(nodes, node)
		}
	}

	return nodes
}
