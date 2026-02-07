package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/service"
)

// SearchHandler handles search operations
type SearchHandler struct {
	searchService *service.SearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchRequest represents search request parameters
type SearchRequest struct {
	Query      string `json:"q" form:"q"`
	Type       []int  `json:"type" form:"type"`
	CategoryID uint   `json:"category_id" form:"category_id"`
	Status     []int  `json:"status" form:"status"`
	Level      int    `json:"level" form:"level"`
	DateFrom   string `json:"date_from" form:"date_from"`
	DateTo     string `json:"date_to" form:"date_to"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	SortBy     string `json:"sort_by" form:"sort_by"` // relevance, date, title
}

// Search handles search requests
func (h *SearchHandler) Search() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid search parameters",
				"error":   err.Error(),
			})
			return
		}

		// Set defaults
		if req.Page < 1 {
			req.Page = 1
		}
		if req.PageSize < 1 || req.PageSize > 100 {
			req.PageSize = 20
		}
		if req.SortBy == "" {
			req.SortBy = "relevance"
		}

		// Get user context for access control
		_, _ = c.Get("user_id")
		groupID, _ := c.Get("group_id")
		userLevel, _ := c.Get("user_level")

		// Build search params
		params := service.SearchParams{
			Query:      req.Query,
			Type:       req.Type,
			CategoryID: req.CategoryID,
			Status:     req.Status,
			Level:      req.Level,
			DateFrom:   req.DateFrom,
			DateTo:     req.DateTo,
			Page:       req.Page,
			PageSize:   req.PageSize,
			SortBy:     req.SortBy,
		}

		// Apply access control
		if groupID != nil {
			if gid, ok := groupID.(uint); ok {
				params.GroupID = gid
			}
		}
		if userLevel != nil {
			if level, ok := userLevel.(int); ok {
				params.Level = level
			}
		}

		// Perform search
		results, total, facets, err := h.searchService.Search(
			c.Request.Context(),
			params,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Search failed",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"results": results,
			"pagination": gin.H{
				"page":        req.Page,
				"page_size":   req.PageSize,
				"total":       total,
				"total_pages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
			},
			"facets": facets,
			"query":  req.Query,
		})
	}
}

// Reindex triggers search index rebuild (admin only)
func (h *SearchHandler) Reindex() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start reindex process
		if err := h.searchService.Reindex(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to trigger reindex",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Reindex started successfully",
		})
	}
}

// GetIndexStatus returns indexing status
func (h *SearchHandler) GetIndexStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := h.searchService.GetIndexStatus(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get index status",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  status,
		})
	}
}


// GetSuggestions returns search suggestions for autocomplete
func (h *SearchHandler) GetSuggestions() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Query parameter 'q' is required",
			})
			return
		}

		// Get suggestions (popular/recent search terms)
		suggestions, err := h.searchService.GetSuggestions(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get search suggestions",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"suggestions": suggestions,
		})
	}
}

