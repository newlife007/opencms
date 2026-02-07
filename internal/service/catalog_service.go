package service

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// CatalogService handles catalog metadata configuration
type CatalogService struct {
	db *gorm.DB
}

// NewCatalogService creates a new catalog service
func NewCatalogService(db *gorm.DB) *CatalogService {
	return &CatalogService{db: db}
}

// GetCatalogByType retrieves catalog configuration for a file type
func (s *CatalogService) GetCatalogByType(ctx context.Context, fileType int) ([]models.Catalog, error) {
	var catalogs []models.Catalog
	err := s.db.WithContext(ctx).Where("enabled = ?", true).
		Order("weight ASC").Find(&catalogs).Error
	return catalogs, err
}

// GetAllCatalogs retrieves all catalog configurations
func (s *CatalogService) GetAllCatalogs(ctx context.Context) ([]models.Catalog, error) {
	var catalogs []models.Catalog
	err := s.db.WithContext(ctx).Order("weight ASC").Find(&catalogs).Error
	return catalogs, err
}

// GetCatalogByID retrieves a catalog by ID
func (s *CatalogService) GetCatalogByID(ctx context.Context, id uint) (*models.Catalog, error) {
	var catalog models.Catalog
	err := s.db.WithContext(ctx).First(&catalog, id).Error
	if err != nil {
		return nil, err
	}
	return &catalog, nil
}

// CreateCatalog creates a new catalog configuration
func (s *CatalogService) CreateCatalog(ctx context.Context, catalog *models.Catalog) error {
	return s.db.WithContext(ctx).Create(catalog).Error
}

// UpdateCatalog updates catalog configuration
func (s *CatalogService) UpdateCatalog(ctx context.Context, id uint, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Model(&models.Catalog{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteCatalog deletes a catalog configuration
func (s *CatalogService) DeleteCatalog(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Catalog{}, id).Error
}

// CatalogNode represents a catalog tree node
type CatalogNode struct {
	ID          int            `json:"id"`
	Type        int            `json:"type"` // File type
	ParentID    int            `json:"parent_id"`
	Path        string         `json:"path"`
	Name        string         `json:"name"`
	Label       string         `json:"label"` // Display label
	Description string         `json:"description"`
	FieldType   string         `json:"field_type"` // text, number, date, select, textarea
	Required    bool           `json:"required"`
	Options     string         `json:"options"` // JSON for select options
	Weight      int            `json:"weight"`
	Enabled     bool           `json:"enabled"`
	Children    []CatalogNode  `json:"children,omitempty"`
}

// GetCatalogTree builds hierarchical tree structure filtered by file type
func (s *CatalogService) GetCatalogTree(ctx context.Context, fileType int) ([]CatalogNode, error) {
	var catalogs []models.Catalog
	
	// Query catalogs by type field
	err := s.db.WithContext(ctx).
		Where("type = ? AND enabled = ?", fileType, true).
		Order("weight ASC, id ASC").
		Find(&catalogs).Error
	if err != nil {
		return nil, err
	}
	
	// Build flat map of all nodes first
	nodeMap := make(map[int]*CatalogNode)
	for i := range catalogs {
		cat := &catalogs[i]
		node := &CatalogNode{
			ID:          cat.ID,
			Type:        cat.Type,
			ParentID:    cat.ParentID,
			Path:        cat.Path,
			Name:        cat.Name,
			Label:       cat.Label,
			Description: cat.Description,
			FieldType:   cat.FieldType,
			Required:    cat.Required,
			Options:     cat.Options,
			Weight:      cat.Weight,
			Enabled:     cat.Enabled,
			Children:    make([]CatalogNode, 0),
		}
		nodeMap[cat.ID] = node
	}
	
	// Build tree by linking children to parents
	// We must modify the pointers in nodeMap, not copies
	for _, node := range nodeMap {
		if node.ParentID != 0 {
			if parent, exists := nodeMap[node.ParentID]; exists {
				// Append pointer to children (not value!)
				// We'll convert to values at the end
				parent.Children = append(parent.Children, *node)
			}
		}
	}
	
	// Now recursively resolve children for each node
	// This ensures all grandchildren are properly included
	var resolveChildren func(*CatalogNode)
	resolveChildren = func(n *CatalogNode) {
		// For each child in this node's Children slice
		for i := range n.Children {
			child := &n.Children[i]
			// Get the latest version from nodeMap (which has Children populated)
			if latestChild, exists := nodeMap[child.ID]; exists {
				// Copy the Children from the latest version
				child.Children = make([]CatalogNode, len(latestChild.Children))
				copy(child.Children, latestChild.Children)
				// Recursively resolve grandchildren
				resolveChildren(child)
			}
		}
	}
	
	// Collect root nodes and resolve their trees
	var result []CatalogNode
	for _, node := range nodeMap {
		if node.ParentID == 0 {
			// This is a root node
			// Resolve all its descendants
			resolveChildren(node)
			// Add to result
			result = append(result, *node)
		}
	}
	
	return result, nil
}
