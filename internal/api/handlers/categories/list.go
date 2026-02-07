package categories

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListHandler handles category tree listing
type ListHandler struct{}

func NewListHandler() *ListHandler {
	return &ListHandler{}
}

func (h *ListHandler) GetCategoryTree(c *gin.Context) {
	// TODO: Build category tree from database
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}
