package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/service"
)

// CatalogHandler handles catalog metadata configuration
type CatalogHandler struct {
	catalogService *service.CatalogService
}

// NewCatalogHandler creates a new catalog handler
func NewCatalogHandler(catalogService *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{
		catalogService: catalogService,
	}
}

// GetCatalogConfig returns catalog configuration for a file type
func (h *CatalogHandler) GetCatalogConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileType, _ := strconv.Atoi(c.Query("type"))
		
		if fileType <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "File type parameter is required",
			})
			return
		}

		// Get hierarchical catalog tree
		tree, err := h.catalogService.GetCatalogTree(c.Request.Context(), fileType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve catalog configuration",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"type":    fileType,
			"catalog": tree,
		})
	}
}

// ListCatalogs returns all catalog configurations

// GetCatalogTree returns hierarchical catalog tree by file type
// This is an alias for GetCatalogConfig for frontend compatibility
func (h *CatalogHandler) GetCatalogTree() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileType, _ := strconv.Atoi(c.Query("type"))
		
		if fileType <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "File type parameter is required",
			})
			return
		}

		// Get hierarchical catalog tree
		tree, err := h.catalogService.GetCatalogTree(c.Request.Context(), fileType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve catalog tree",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    tree,
		})
	}
}

func (h *CatalogHandler) ListCatalogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		catalogs, err := h.catalogService.GetAllCatalogs(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve catalogs",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    catalogs,
		})
	}
}

// GetCatalog returns catalog details by ID
func (h *CatalogHandler) GetCatalog() gin.HandlerFunc {
	return func(c *gin.Context) {
		catalogID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid catalog ID",
			})
			return
		}

		catalog, err := h.catalogService.GetCatalogByID(c.Request.Context(), uint(catalogID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Catalog not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    catalog,
		})
	}
}

// CreateCatalog creates a new catalog configuration
func (h *CatalogHandler) CreateCatalog() gin.HandlerFunc {
	return func(c *gin.Context) {
		var catalog models.Catalog
		if err := c.ShouldBindJSON(&catalog); err != nil {
			// Log the error with more details
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body - failed to parse JSON or validate fields",
				"error":   err.Error(),
			})
			return
		}

		if err := h.catalogService.CreateCatalog(c.Request.Context(), &catalog); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create catalog in database",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Catalog created successfully",
			"data":    catalog,
		})
	}
}

// UpdateCatalog updates catalog configuration
func (h *CatalogHandler) UpdateCatalog() gin.HandlerFunc {
	return func(c *gin.Context) {
		catalogID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid catalog ID",
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

		if err := h.catalogService.UpdateCatalog(c.Request.Context(), uint(catalogID), updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update catalog",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Catalog updated successfully",
		})
	}
}

// DeleteCatalog deletes a catalog configuration
func (h *CatalogHandler) DeleteCatalog() gin.HandlerFunc {
	return func(c *gin.Context) {
		catalogID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid catalog ID",
			})
			return
		}

		if err := h.catalogService.DeleteCatalog(c.Request.Context(), uint(catalogID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete catalog",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Catalog deleted successfully",
		})
	}
}
