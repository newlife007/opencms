package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// permissionsRepository implements PermissionsRepository
type permissionsRepository struct {
	db *gorm.DB
}

// NewPermissionsRepository creates a new permissions repository
func NewPermissionsRepository(db *gorm.DB) PermissionsRepository {
	return &permissionsRepository{db: db}
}

func (r *permissionsRepository) Create(ctx context.Context, permission *models.Permissions) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionsRepository) FindByID(ctx context.Context, id int) (*models.Permissions, error) {
	var permission models.Permissions
	err := r.db.WithContext(ctx).First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionsRepository) FindAll(ctx context.Context) ([]*models.Permissions, error) {
	var permissions []*models.Permissions
	err := r.db.WithContext(ctx).Order("namespace ASC, controller ASC, action ASC").Find(&permissions).Error
	return permissions, err
}

func (r *permissionsRepository) Update(ctx context.Context, permission *models.Permissions) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *permissionsRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Permissions{}, id).Error
}

func (r *permissionsRepository) FindByNamespaceController(ctx context.Context, namespace, controller, action string) (*models.Permissions, error) {
	var permission models.Permissions
	err := r.db.WithContext(ctx).
		Where("namespace = ? AND controller = ? AND action = ?", namespace, controller, action).
		First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// Permissions returns permissions repository implementation
func Permissions(db *gorm.DB) PermissionsRepository {
	return NewPermissionsRepository(db)
}
