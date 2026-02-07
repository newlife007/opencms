package models

// Groups represents the ow_groups table for user group management
type Groups struct {
	ID          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(32);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Quota       int    `gorm:"column:quota;not null;default:1000" json:"quota"` // User disk quota in MB
	Weight      int    `gorm:"column:weight;not null;default:0" json:"weight"`
	Enabled     bool   `gorm:"column:enabled;type:tinyint(2);not null;default:true" json:"enabled"`

	// Many-to-many relationships
	Roles      []Roles     `gorm:"many2many:ow_groups_has_roles;foreignKey:ID;joinForeignKey:group_id;References:ID;joinReferences:role_id" json:"roles,omitempty"`
	Categories []Category `gorm:"many2many:ow_groups_has_category;foreignKey:ID;joinForeignKey:group_id;References:ID;joinReferences:category_id" json:"categories,omitempty"`
}

// TableName specifies the table name for the Groups model
func (Groups) TableName() string {
	return "ow_groups"
}
