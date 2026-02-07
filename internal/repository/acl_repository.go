package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// aclRepository implements ACLRepository
type aclRepository struct {
	db *gorm.DB
}

// NewACLRepository creates a new ACL repository
func NewACLRepository(db *gorm.DB) ACLRepository {
	return &aclRepository{db: db}
}

func (r *aclRepository) HasPermission(ctx context.Context, userID int, namespace, controller, action string) (bool, error) {
	// Get user with group
	var user models.Users
	if err := r.db.WithContext(ctx).Preload("Group").First(&user, userID).Error; err != nil {
		return false, err
	}

	// Check if user's group has roles with this permission
	var count int64
	err := r.db.WithContext(ctx).
		Table("ow_permissions p").
		Joins("JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id").
		Joins("JOIN ow_groups_has_roles ghr ON rhp.role_id = ghr.role_id").
		Where("ghr.group_id = ? AND p.namespace = ? AND p.controller = ? AND p.action = ?",
			user.GroupID, namespace, controller, action).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *aclRepository) GetUserPermissions(ctx context.Context, userID int) ([]*models.Permissions, error) {
	var user models.Users
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	var permissions []*models.Permissions
	err := r.db.WithContext(ctx).
		Table("ow_permissions p").
		Joins("JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id").
		Joins("JOIN ow_groups_has_roles ghr ON rhp.role_id = ghr.role_id").
		Where("ghr.group_id = ?", user.GroupID).
		Find(&permissions).Error

	return permissions, err
}

func (r *aclRepository) GetUserRoles(ctx context.Context, userID int) ([]*models.Roles, error) {
	var user models.Users
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	var roles []*models.Roles
	err := r.db.WithContext(ctx).
		Table("ow_roles r").
		Joins("JOIN ow_groups_has_roles ghr ON r.id = ghr.role_id").
		Where("ghr.group_id = ?", user.GroupID).
		Find(&roles).Error

	return roles, err
}

func (r *aclRepository) CanAccessCategory(ctx context.Context, userID, categoryID int) (bool, error) {
	var user models.Users
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return false, err
	}

	// Check if user's group has access to this category
	var count int64
	err := r.db.WithContext(ctx).
		Table("ow_groups_has_category").
		Where("group_id = ? AND category_id = ?", user.GroupID, categoryID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *aclRepository) CanAccessFile(ctx context.Context, userID int, fileID uint64) (bool, error) {
	var user models.Users
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return false, err
	}

	// Get user's level details
	var userLevel models.Levels
	if err := r.db.WithContext(ctx).First(&userLevel, user.LevelID).Error; err != nil {
		return false, err
	}

	var file models.Files
	if err := r.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
		return false, err
	}

	// Check level access
	// Logic: User can only view files with level <= user's level
	// Higher level = More access (e.g., level 5 user can see level 1,2,3,4,5 files)
	if file.Level > userLevel.Level {
		return false, nil
	}

	// Check group access
	if file.Groups == "all" {
		return true, nil
	}

	// Parse comma-separated group IDs
	allowedGroups := strings.Split(file.Groups, ",")
	userGroupStr := fmt.Sprintf("%d", user.GroupID)
	for _, g := range allowedGroups {
		if strings.TrimSpace(g) == userGroupStr {
			return true, nil
		}
	}

	return false, nil
}

func (r *aclRepository) IsAdmin(ctx context.Context, userID int) (bool, error) {
	roles, err := r.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == models.RoleAdmin || role.Name == models.RoleSystem {
			return true, nil
		}
	}

	return false, nil
}
