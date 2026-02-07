package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// categoryRepository implements CategoryRepository
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) FindByID(ctx context.Context, id int) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindAll(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("enabled = ?", 1).
		Order("weight ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) FindByParentID(ctx context.Context, parentID int) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("parent_id = ? AND enabled = ?", parentID, 1).
		Order("weight ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	// Soft delete by setting enabled to 0
	return r.db.WithContext(ctx).Model(&models.Category{}).
		Where("id = ?", id).
		Update("enabled", 0).Error
}

func (r *categoryRepository) BuildTree(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("enabled = ?", 1).
		Order("weight ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) FindAccessibleByGroupID(ctx context.Context, groupID int) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Table("ow_category c").
		Joins("JOIN ow_groups_has_category ghc ON c.id = ghc.category_id").
		Where("ghc.group_id = ? AND c.enabled = ?", groupID, 1).
		Order("c.weight ASC, c.id ASC").
		Find(&categories).Error
	return categories, err
}

// Categories returns category repository implementation
func Categories(db *gorm.DB) CategoryRepository {
	return NewCategoryRepository(db)
}
