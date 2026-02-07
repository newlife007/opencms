package repository

import (
	"context"

	"gorm.io/gorm"
)

// repository implements Repository interface
type repository struct {
	db                *gorm.DB
	filesRepo         FilesRepository
	catalogRepo       CatalogRepository
	categoryRepo      CategoryRepository
	usersRepo         UsersRepository
	groupsRepo        GroupsRepository
	rolesRepo         RolesRepository
	permissionsRepo   PermissionsRepository
	levelsRepo        LevelsRepository
	aclRepo           ACLRepository
}

// NewRepository creates a new repository factory
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db:              db,
		filesRepo:       NewFilesRepository(db),
		catalogRepo:     NewCatalogRepository(db),
		categoryRepo:    NewCategoryRepository(db),
		usersRepo:       NewUsersRepository(db),
		groupsRepo:      NewGroupsRepository(db),
		rolesRepo:       NewRolesRepository(db),
		permissionsRepo: NewPermissionsRepository(db),
		levelsRepo:      NewLevelsRepository(db),
		aclRepo:         NewACLRepository(db),
	}
}

func (r *repository) Files() FilesRepository {
	return r.filesRepo
}

func (r *repository) Catalog() CatalogRepository {
	return r.catalogRepo
}

func (r *repository) Category() CategoryRepository {
	return r.categoryRepo
}

func (r *repository) Users() UsersRepository {
	return r.usersRepo
}

func (r *repository) Groups() GroupsRepository {
	return r.groupsRepo
}

func (r *repository) Roles() RolesRepository {
	return r.rolesRepo
}

func (r *repository) Permissions() PermissionsRepository {
	return r.permissionsRepo
}

func (r *repository) Levels() LevelsRepository {
	return r.levelsRepo
}

func (r *repository) ACL() ACLRepository {
	return r.aclRepo
}

func (r *repository) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}
