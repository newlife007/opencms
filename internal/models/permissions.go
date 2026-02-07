package models

// Permissions represents the ow_permissions table for access control permissions
type Permissions struct {
	ID         int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Namespace  string `gorm:"column:namespace;type:varchar(64);not null;default:'default'" json:"namespace"`
	Controller string `gorm:"column:controller;type:varchar(64);not null;default:'default'" json:"controller"`
	Action     string `gorm:"column:action;type:varchar(64);not null;default:'index'" json:"action"`
	Aliasname  string `gorm:"column:aliasname;type:varchar(64);not null;default:''" json:"aliasname"`
	RBAC       string `gorm:"column:rbac;type:varchar(32);not null;default:'ACL_NULL'" json:"rbac"`
}

// TableName specifies the table name for the Permissions model
func (Permissions) TableName() string {
	return "ow_permissions"
}

// RBAC constants
const (
	RBACNull     = "ACL_NULL"
	RBACEveryone = "ACL_EVERYONE"
	RBACNoRole   = "ACL_NO_ROLE"
	RBACHasRole  = "ACL_HAS_ROLE"
)
