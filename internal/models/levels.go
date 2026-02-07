package models

// Levels represents the ow_levels table for browsing level (access level) management
// Level logic: User can only view files with level <= user's level
// Higher level = More access (e.g., level 5 user can see level 1,2,3,4,5 files)
type Levels struct {
	ID          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Level       int    `gorm:"column:level;not null;default:1" json:"level"` // Browsing level value (1-10)
	Enabled     bool   `gorm:"column:enabled;type:tinyint(2);not null;default:true" json:"enabled"`
}

// TableName specifies the table name for the Levels model
func (Levels) TableName() string {
	return "ow_levels"
}
