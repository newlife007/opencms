package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// groupsRepository implements GroupsRepository
type groupsRepository struct {
	db *gorm.DB
}

// NewGroupsRepository creates a new groups repository
func NewGroupsRepository(db *gorm.DB) GroupsRepository {
	return &groupsRepository{db: db}
}

func (r *groupsRepository) Create(ctx context.Context, group *models.Groups) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *groupsRepository) FindByID(ctx context.Context, id int) (*models.Groups, error) {
	var group models.Groups
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupsRepository) FindAll(ctx context.Context) ([]*models.Groups, error) {
	var groups []*models.Groups
	err := r.db.WithContext(ctx).Find(&groups).Error
	return groups, err
}

func (r *groupsRepository) Update(ctx context.Context, group *models.Groups) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *groupsRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Groups{}, id).Error
}

func (r *groupsRepository) AssignRole(ctx context.Context, groupID, roleID int) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO ow_groups_has_roles (group_id, role_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE group_id=group_id",
		groupID, roleID,
	).Error
}

func (r *groupsRepository) RemoveRole(ctx context.Context, groupID, roleID int) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM ow_groups_has_roles WHERE group_id = ? AND role_id = ?",
		groupID, roleID,
	).Error
}

func (r *groupsRepository) GetRoles(ctx context.Context, groupID int) ([]*models.Roles, error) {
	var roles []*models.Roles
	err := r.db.WithContext(ctx).
		Table("ow_roles r").
		Joins("JOIN ow_groups_has_roles ghr ON r.id = ghr.role_id").
		Where("ghr.group_id = ?", groupID).
		Find(&roles).Error
	return roles, err
}

func (r *groupsRepository) AssignCategory(ctx context.Context, groupID, categoryID int) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO ow_groups_has_category (group_id, category_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE group_id=group_id",
		groupID, categoryID,
	).Error
}

func (r *groupsRepository) RemoveCategory(ctx context.Context, groupID, categoryID int) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM ow_groups_has_category WHERE group_id = ? AND category_id = ?",
		groupID, categoryID,
	).Error
}

// Groups returns groups repository implementation
func Groups(db *gorm.DB) GroupsRepository {
	return NewGroupsRepository(db)
}
