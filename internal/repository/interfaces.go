package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// FilesRepository interface defines methods for Files data access
type FilesRepository interface {
	Create(ctx context.Context, file *models.Files) error
	FindByID(ctx context.Context, id uint64) (*models.Files, error)
	FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Files, int64, error)
	Update(ctx context.Context, file *models.Files) error
	Delete(ctx context.Context, id uint64) error
	FindByMD5(ctx context.Context, md5 string) (*models.Files, error)
	FindByStatusAndType(ctx context.Context, status, fileType int, limit, offset int) ([]*models.Files, int64, error)
	UpdateStatus(ctx context.Context, id uint64, status int, username string) error
}

// CatalogRepository interface for Catalog data access
type CatalogRepository interface {
	Create(ctx context.Context, catalog *models.Catalog) error
	FindByID(ctx context.Context, id int) (*models.Catalog, error)
	FindAll(ctx context.Context) ([]*models.Catalog, error)
	FindByParentID(ctx context.Context, parentID int) ([]*models.Catalog, error)
	FindByPath(ctx context.Context, pathPrefix string) ([]*models.Catalog, error)
	Update(ctx context.Context, catalog *models.Catalog) error
	Delete(ctx context.Context, id int) error
	BuildTree(ctx context.Context) ([]*models.Catalog, error)
}

// CategoryRepository interface for Category data access
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	FindByID(ctx context.Context, id int) (*models.Category, error)
	FindAll(ctx context.Context) ([]*models.Category, error)
	FindByParentID(ctx context.Context, parentID int) ([]*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id int) error
	BuildTree(ctx context.Context) ([]*models.Category, error)
	FindAccessibleByGroupID(ctx context.Context, groupID int) ([]*models.Category, error)
}

// UsersRepository interface for Users data access
type UsersRepository interface {
	Create(ctx context.Context, user *models.Users) error
	FindByID(ctx context.Context, id int) (*models.Users, error)
	GetByID(ctx context.Context, id uint) (*models.Users, error) // With full relationships
	FindByUsername(ctx context.Context, username string) (*models.Users, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.Users, int64, error)
	Update(ctx context.Context, user *models.Users) error
	Delete(ctx context.Context, id int) error
	UpdatePassword(ctx context.Context, id int, hashedPassword string) error
	UpdateStatus(ctx context.Context, id uint, status int) error
	UpdateLoginInfo(ctx context.Context, id int, ip string) error
	SearchUsers(ctx context.Context, params UserSearchParams) ([]*models.Users, int64, error)
	BatchDelete(ctx context.Context, ids []int) error
	CheckUsernameExists(ctx context.Context, username string, excludeID int) (bool, error)
}

// GroupsRepository interface for Groups data access
type GroupsRepository interface {
	Create(ctx context.Context, group *models.Groups) error
	FindByID(ctx context.Context, id int) (*models.Groups, error)
	FindAll(ctx context.Context) ([]*models.Groups, error)
	Update(ctx context.Context, group *models.Groups) error
	Delete(ctx context.Context, id int) error
	AssignRole(ctx context.Context, groupID, roleID int) error
	RemoveRole(ctx context.Context, groupID, roleID int) error
	GetRoles(ctx context.Context, groupID int) ([]*models.Roles, error)
	AssignCategory(ctx context.Context, groupID, categoryID int) error
	RemoveCategory(ctx context.Context, groupID, categoryID int) error
}

// RolesRepository interface for Roles data access
type RolesRepository interface {
	Create(ctx context.Context, role *models.Roles) error
	FindByID(ctx context.Context, id int) (*models.Roles, error)
	FindAll(ctx context.Context) ([]*models.Roles, error)
	FindByName(ctx context.Context, name string) (*models.Roles, error)
	Update(ctx context.Context, role *models.Roles) error
	Delete(ctx context.Context, id int) error
	AssignPermission(ctx context.Context, roleID, permissionID int) error
	RemovePermission(ctx context.Context, roleID, permissionID int) error
	ClearPermissions(ctx context.Context, roleID int) error
	GetPermissions(ctx context.Context, roleID int) ([]*models.Permissions, error)
}

// PermissionsRepository interface for Permissions data access
type PermissionsRepository interface {
	Create(ctx context.Context, permission *models.Permissions) error
	FindByID(ctx context.Context, id int) (*models.Permissions, error)
	FindAll(ctx context.Context) ([]*models.Permissions, error)
	Update(ctx context.Context, permission *models.Permissions) error
	Delete(ctx context.Context, id int) error
	FindByNamespaceController(ctx context.Context, namespace, controller, action string) (*models.Permissions, error)
}

// LevelsRepository interface for Levels data access
type LevelsRepository interface {
	Create(ctx context.Context, level *models.Levels) error
	FindByID(ctx context.Context, id int) (*models.Levels, error)
	FindAll(ctx context.Context) ([]*models.Levels, error)
	Update(ctx context.Context, level *models.Levels) error
	Delete(ctx context.Context, id int) error
}

// ACLRepository interface for RBAC permission checking
type ACLRepository interface {
	HasPermission(ctx context.Context, userID int, namespace, controller, action string) (bool, error)
	GetUserPermissions(ctx context.Context, userID int) ([]*models.Permissions, error)
	GetUserRoles(ctx context.Context, userID int) ([]*models.Roles, error)
	CanAccessCategory(ctx context.Context, userID, categoryID int) (bool, error)
	CanAccessFile(ctx context.Context, userID int, fileID uint64) (bool, error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
}

// TransactionFunc is a function type for transaction operations
type TransactionFunc func(ctx context.Context, tx *gorm.DB) error

// Repository interface combines all repositories
type Repository interface {
	Files() FilesRepository
	Catalog() CatalogRepository
	Category() CategoryRepository
	Users() UsersRepository
	Groups() GroupsRepository
	Roles() RolesRepository
	Permissions() PermissionsRepository
	Levels() LevelsRepository
	ACL() ACLRepository
	
	// Transaction support
	WithTransaction(ctx context.Context, fn TransactionFunc) error
}
