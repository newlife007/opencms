package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// catalogRepository implements CatalogRepository
type catalogRepository struct {
	db *gorm.DB
}

// NewCatalogRepository creates a new catalog repository
func NewCatalogRepository(db *gorm.DB) CatalogRepository {
	return &catalogRepository{db: db}
}

func (r *catalogRepository) Create(ctx context.Context, catalog *models.Catalog) error {
	return r.db.WithContext(ctx).Create(catalog).Error
}

func (r *catalogRepository) FindByID(ctx context.Context, id int) (*models.Catalog, error) {
	var catalog models.Catalog
	err := r.db.WithContext(ctx).First(&catalog, id).Error
	if err != nil {
		return nil, err
	}
	return &catalog, nil
}

func (r *catalogRepository) FindAll(ctx context.Context) ([]*models.Catalog, error) {
	var catalogs []*models.Catalog
	err := r.db.WithContext(ctx).Order("weight ASC, id ASC").Find(&catalogs).Error
	return catalogs, err
}

func (r *catalogRepository) FindByParentID(ctx context.Context, parentID int) ([]*models.Catalog, error) {
	var catalogs []*models.Catalog
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("weight ASC, id ASC").
		Find(&catalogs).Error
	return catalogs, err
}

func (r *catalogRepository) FindByPath(ctx context.Context, pathPrefix string) ([]*models.Catalog, error) {
	var catalogs []*models.Catalog
	err := r.db.WithContext(ctx).
		Where("path LIKE ?", pathPrefix+"%").
		Order("weight ASC, id ASC").
		Find(&catalogs).Error
	return catalogs, err
}

func (r *catalogRepository) Update(ctx context.Context, catalog *models.Catalog) error {
	return r.db.WithContext(ctx).Save(catalog).Error
}

func (r *catalogRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Catalog{}, id).Error
}

func (r *catalogRepository) BuildTree(ctx context.Context) ([]*models.Catalog, error) {
	var catalogs []*models.Catalog
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("weight ASC, id ASC").
		Find(&catalogs).Error
	return catalogs, err
}

// Catalogs returns catalog repository implementation
func Catalogs(db *gorm.DB) CatalogRepository {
	return NewCatalogRepository(db)
}
