package service

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/repository"
)

// GroupService handles group-related business logic
type GroupService struct {
	repo repository.Repository
}

// NewGroupService creates a new group service
func NewGroupService(repo repository.Repository) *GroupService {
	return &GroupService{repo: repo}
}

// ListGroups retrieves all groups with pagination
func (s *GroupService) ListGroups(ctx context.Context, limit, offset int) ([]*models.Groups, int64, error) {
	groups, err := s.repo.Groups().FindAll(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	// Calculate pagination
	total := int64(len(groups))
	start := offset
	end := offset + limit
	if start > len(groups) {
		start = len(groups)
	}
	if end > len(groups) {
		end = len(groups)
	}
	
	return groups[start:end], total, nil
}

// GetGroupByID retrieves a group by ID
func (s *GroupService) GetGroupByID(ctx context.Context, id uint) (*models.Groups, error) {
	return s.repo.Groups().FindByID(ctx, int(id))
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(ctx context.Context, group *models.Groups) error {
	return s.repo.Groups().Create(ctx, group)
}

// UpdateGroup updates group information
func (s *GroupService) UpdateGroup(ctx context.Context, id uint, updates map[string]interface{}) error {
	group, err := s.repo.Groups().FindByID(ctx, int(id))
	if err != nil {
		return err
	}
	
	// Apply updates
	if name, ok := updates["name"].(string); ok {
		group.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		group.Description = description
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		group.Enabled = enabled
	}
	
	return s.repo.Groups().Update(ctx, group)
}

// DeleteGroup deletes a group
func (s *GroupService) DeleteGroup(ctx context.Context, id uint) error {
	return s.repo.Groups().Delete(ctx, int(id))
}

// AssignCategories assigns categories to a group
func (s *GroupService) AssignCategories(ctx context.Context, groupID uint, categoryIDs []uint) error {
	for _, categoryID := range categoryIDs {
		if err := s.repo.Groups().AssignCategory(ctx, int(groupID), int(categoryID)); err != nil {
			return err
		}
	}
	return nil
}

// GetGroupCategories retrieves categories assigned to a group
func (s *GroupService) GetGroupCategories(ctx context.Context, groupID uint) ([]models.Category, error) {
	categories, err := s.repo.Category().FindAccessibleByGroupID(ctx, int(groupID))
	if err != nil {
		return nil, err
	}
	
	result := make([]models.Category, len(categories))
	for i, cat := range categories {
		result[i] = *cat
	}
	return result, nil
}

// AssignRoles assigns roles to a group
func (s *GroupService) AssignRoles(ctx context.Context, groupID uint, roleIDs []uint) error {
	for _, roleID := range roleIDs {
		if err := s.repo.Groups().AssignRole(ctx, int(groupID), int(roleID)); err != nil {
			return err
		}
	}
	return nil
}

// GetGroupRoles retrieves roles assigned to a group
func (s *GroupService) GetGroupRoles(ctx context.Context, groupID uint) ([]models.Roles, error) {
	roles, err := s.repo.Groups().GetRoles(ctx, int(groupID))
	if err != nil {
		return nil, err
	}
	
	result := make([]models.Roles, len(roles))
	for i, role := range roles {
		result[i] = *role
	}
	return result, nil
}

// RoleService handles role-related business logic
type RoleService struct {
	repo repository.Repository
}

// NewRoleService creates a new role service
func NewRoleService(repo repository.Repository) *RoleService {
	return &RoleService{repo: repo}
}

// ListRoles retrieves all roles with pagination
func (s *RoleService) ListRoles(ctx context.Context, limit, offset int) ([]*models.Roles, int64, error) {
	roles, err := s.repo.Roles().FindAll(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	// Calculate pagination
	total := int64(len(roles))
	start := offset
	end := offset + limit
	if start > len(roles) {
		start = len(roles)
	}
	if end > len(roles) {
		end = len(roles)
	}
	
	return roles[start:end], total, nil
}

// GetRoleByID retrieves a role by ID
func (s *RoleService) GetRoleByID(ctx context.Context, id uint) (*models.Roles, error) {
	return s.repo.Roles().FindByID(ctx, int(id))
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, role *models.Roles) error {
	return s.repo.Roles().Create(ctx, role)
}

// UpdateRole updates role information
func (s *RoleService) UpdateRole(ctx context.Context, id uint, updates map[string]interface{}) error {
	role, err := s.repo.Roles().FindByID(ctx, int(id))
	if err != nil {
		return err
	}
	
	// Apply updates
	if name, ok := updates["name"].(string); ok {
		role.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		role.Description = description
	}
	
	return s.repo.Roles().Update(ctx, role)
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, id uint) error {
	return s.repo.Roles().Delete(ctx, int(id))
}

// AssignPermissions assigns permissions to a role
// This replaces all existing permissions with the new set
func (s *RoleService) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	// First, remove all existing permissions
	if err := s.repo.Roles().ClearPermissions(ctx, int(roleID)); err != nil {
		return err
	}
	
	// Then assign new permissions
	for _, permissionID := range permissionIDs {
		if err := s.repo.Roles().AssignPermission(ctx, int(roleID), int(permissionID)); err != nil {
			return err
		}
	}
	return nil
}

// GetRolePermissions retrieves permissions assigned to a role
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID uint) ([]models.Permissions, error) {
	permissions, err := s.repo.Roles().GetPermissions(ctx, int(roleID))
	if err != nil {
		return nil, err
	}
	
	result := make([]models.Permissions, len(permissions))
	for i, perm := range permissions {
		result[i] = *perm
	}
	return result, nil
}

// PermissionService handles permission-related business logic
type PermissionService struct {
	repo repository.Repository
}

// NewPermissionService creates a new permission service
func NewPermissionService(repo repository.Repository) *PermissionService {
	return &PermissionService{repo: repo}
}

// ListPermissions retrieves all permissions
func (s *PermissionService) ListPermissions(ctx context.Context) ([]*models.Permissions, error) {
	return s.repo.Permissions().FindAll(ctx)
}

// GetPermissionByID retrieves a permission by ID
func (s *PermissionService) GetPermissionByID(ctx context.Context, id uint) (*models.Permissions, error) {
	return s.repo.Permissions().FindByID(ctx, int(id))
}

// CreatePermission creates a new permission
func (s *PermissionService) CreatePermission(ctx context.Context, permission *models.Permissions) error {
	return s.repo.Permissions().Create(ctx, permission)
}

// UpdatePermission updates permission information
func (s *PermissionService) UpdatePermission(ctx context.Context, id uint, updates map[string]interface{}) error {
	permission, err := s.repo.Permissions().FindByID(ctx, int(id))
	if err != nil {
		return err
	}
	
	// Apply updates - Permissions model uses aliasname instead of name/description
	if aliasname, ok := updates["aliasname"].(string); ok {
		permission.Aliasname = aliasname
	}
	if namespace, ok := updates["namespace"].(string); ok {
		permission.Namespace = namespace
	}
	if controller, ok := updates["controller"].(string); ok {
		permission.Controller = controller
	}
	if action, ok := updates["action"].(string); ok {
		permission.Action = action
	}
	
	return s.repo.Permissions().Update(ctx, permission)
}

// DeletePermission deletes a permission
func (s *PermissionService) DeletePermission(ctx context.Context, id uint) error {
	return s.repo.Permissions().Delete(ctx, int(id))
}
