package models

// Category represents the ow_category table for hierarchical resource classification
type Category struct {
	ID          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParentID    int    `gorm:"column:parent_id;not null;index" json:"parent_id"`
	Path        string `gorm:"column:path;type:varchar(255);not null;index" json:"path"` // hierarchical path like "-1,1,2,"
	Name        string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Weight      int    `gorm:"column:weight;not null;default:0" json:"weight"`
	Enabled     bool   `gorm:"column:enabled;type:tinyint(2);not null;default:true" json:"enabled"`
	Created     int    `gorm:"column:created;not null" json:"created"` // Unix timestamp
	Updated     int    `gorm:"column:updated;not null" json:"updated"` // Unix timestamp
}

// TableName specifies the table name for the Category model
func (Category) TableName() string {
	return "ow_category"
}
