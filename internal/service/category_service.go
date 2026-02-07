package service

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/repository"
)

// CategoryService handles category operations
type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// GetAllCategories retrieves all categories
func (s *CategoryService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert []*models.Category to []models.Category
	result := make([]models.Category, len(categories))
	for i, cat := range categories {
		result[i] = *cat
	}
	return result, nil
}

// GetCategoryByID retrieves a category by ID
func (s *CategoryService) GetCategoryByID(ctx context.Context, id uint) (*models.Category, error) {
	return s.categoryRepo.FindByID(ctx, int(id))
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, category *models.Category) error {
	return s.categoryRepo.Create(ctx, category)
}

// UpdateCategory updates category information
func (s *CategoryService) UpdateCategory(ctx context.Context, id uint, updates map[string]interface{}) error {
	category, err := s.categoryRepo.FindByID(ctx, int(id))
	if err != nil {
		return err
	}
	
	// Update fields
	if name, ok := updates["name"].(string); ok {
		category.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		category.Description = description
	}
	if parentID, ok := updates["parent_id"].(float64); ok {
		category.ParentID = int(parentID)
	}
	if weight, ok := updates["weight"].(float64); ok {
		category.Weight = int(weight)
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		category.Enabled = enabled
	}
	
	return s.categoryRepo.Update(ctx, category)
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, id uint) error {
	return s.categoryRepo.Delete(ctx, int(id))
}

// GetCategoriesByParent retrieves categories by parent ID
func (s *CategoryService) GetCategoriesByParent(ctx context.Context, parentID uint) ([]models.Category, error) {
	categories, err := s.categoryRepo.FindByParentID(ctx, int(parentID))
	if err != nil {
		return nil, err
	}
	
	result := make([]models.Category, len(categories))
	for i, cat := range categories {
		result[i] = *cat
	}
	return result, nil
}

// GetAccessibleCategories retrieves categories accessible by a group
func (s *CategoryService) GetAccessibleCategories(ctx context.Context, groupID int) ([]models.Category, error) {
	categories, err := s.categoryRepo.FindAccessibleByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	
	result := make([]models.Category, len(categories))
	for i, cat := range categories {
		result[i] = *cat
	}
	return result, nil
}

// BuildCategoryTree builds hierarchical tree structure
func (s *CategoryService) BuildCategoryTree(ctx context.Context) ([]*models.Category, error) {
	return s.categoryRepo.BuildTree(ctx)
}
