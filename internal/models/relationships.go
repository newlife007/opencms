package models

// GroupsHasCategory represents the ow_groups_has_category junction table
// for many-to-many relationship between groups and categories
type GroupsHasCategory struct {
	GroupID    int `gorm:"column:group_id;primaryKey" json:"group_id"`
	CategoryID int `gorm:"column:category_id;primaryKey" json:"category_id"`
}

// TableName specifies the table name for the GroupsHasCategory model
func (GroupsHasCategory) TableName() string {
	return "ow_groups_has_category"
}

// GroupsHasRoles represents the ow_groups_has_roles junction table
// for many-to-many relationship between groups and roles
type GroupsHasRoles struct {
	GroupID int `gorm:"column:group_id;primaryKey" json:"group_id"`
	RoleID  int `gorm:"column:role_id;primaryKey" json:"role_id"`
}

// TableName specifies the table name for the GroupsHasRoles model
func (GroupsHasRoles) TableName() string {
	return "ow_groups_has_roles"
}

// RolesHasPermissions represents the ow_roles_has_permissions junction table
// for many-to-many relationship between roles and permissions
type RolesHasPermissions struct {
	RoleID       int `gorm:"column:role_id;primaryKey" json:"role_id"`
	PermissionID int `gorm:"column:permission_id;primaryKey" json:"permission_id"`
}

// TableName specifies the table name for the RolesHasPermissions model
func (RolesHasPermissions) TableName() string {
	return "ow_roles_has_permissions"
}

// FilesCounter represents the ow_files_counter table
// for tracking file IDs for Sphinx search indexing
type FilesCounter struct {
	ID     int `gorm:"column:id;primaryKey" json:"id"`
	FileID int `gorm:"column:file_id;not null" json:"file_id"`
}

// TableName specifies the table name for the FilesCounter model
func (FilesCounter) TableName() string {
	return "ow_files_counter"
}
