package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// rolesRepository implements RolesRepository
type rolesRepository struct {
	db *gorm.DB
}

// NewRolesRepository creates a new roles repository
func NewRolesRepository(db *gorm.DB) RolesRepository {
	return &rolesRepository{db: db}
}

func (r *rolesRepository) Create(ctx context.Context, role *models.Roles) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *rolesRepository) FindByID(ctx context.Context, id int) (*models.Roles, error) {
	var role models.Roles
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *rolesRepository) FindAll(ctx context.Context) ([]*models.Roles, error) {
	var roles []*models.Roles
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (r *rolesRepository) FindByName(ctx context.Context, name string) (*models.Roles, error) {
	var role models.Roles
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *rolesRepository) Update(ctx context.Context, role *models.Roles) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *rolesRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Roles{}, id).Error
}

func (r *rolesRepository) AssignPermission(ctx context.Context, roleID, permissionID int) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO ow_roles_has_permissions (role_id, permission_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE role_id=role_id",
		roleID, permissionID,
	).Error
}

func (r *rolesRepository) RemovePermission(ctx context.Context, roleID, permissionID int) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM ow_roles_has_permissions WHERE role_id = ? AND permission_id = ?",
		roleID, permissionID,
	).Error
}

func (r *rolesRepository) ClearPermissions(ctx context.Context, roleID int) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM ow_roles_has_permissions WHERE role_id = ?",
		roleID,
	).Error
}

func (r *rolesRepository) GetPermissions(ctx context.Context, roleID int) ([]*models.Permissions, error) {
	var permissions []*models.Permissions
	err := r.db.WithContext(ctx).
		Table("ow_permissions p").
		Joins("JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id").
		Where("rhp.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

// Roles returns roles repository implementation
func Roles(db *gorm.DB) RolesRepository {
	return NewRolesRepository(db)
}
