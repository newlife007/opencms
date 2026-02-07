package models

// Roles represents the ow_roles table for role-based access control
type Roles struct {
	ID          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(32);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Weight      int    `gorm:"column:weight;not null;default:0" json:"weight"`
	Enabled     bool   `gorm:"column:enabled;type:tinyint(2);not null;default:true" json:"enabled"`
	IsSystem    bool   `gorm:"column:is_system;type:tinyint;not null;default:false" json:"is_system"` // 是否系统角色

	// Many-to-many relationships
	Permissions []Permissions `gorm:"many2many:ow_roles_has_permissions;foreignKey:ID;joinForeignKey:role_id;References:ID;joinReferences:permission_id" json:"permissions,omitempty"`
}

// TableName specifies the table name for the Roles model
func (Roles) TableName() string {
	return "ow_roles"
}

// Role name constants
const (
	RoleAdmin     = "ADMIN"
	RoleSystem    = "SYSTEM"
	RoleNormal    = "NORMAL"
	RoleFreeze    = "FREEZE"
	RoleRepeal    = "REPEAL"
	RoleUnchecked = "UNCHECKED"
)
