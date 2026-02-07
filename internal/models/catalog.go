package models

// Catalog represents the ow_catalog table for dynamic metadata configuration
type Catalog struct {
	ID          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type        int    `gorm:"column:type;not null;default:0;index:idx_type" json:"type"` // File type: 1=video, 2=audio, 3=image, 4=rich
	ParentID    int    `gorm:"column:parent_id;not null;index" json:"parent_id"`
	Path        string `gorm:"column:path;type:varchar(255);not null;index" json:"path"` // hierarchical path like "-1,1,2,"
	Name        string `gorm:"column:name;type:varchar(64);not null" json:"name"`        // Field name (for JSON key)
	Label       string `gorm:"column:label;type:varchar(64);not null;default:''" json:"label"` // Display label
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	FieldType   string `gorm:"column:field_type;type:varchar(32);not null;default:'text'" json:"field_type"` // text, number, date, select, textarea
	Required    bool   `gorm:"column:required;type:tinyint(1);not null;default:false" json:"required"`
	Options     string `gorm:"column:options;type:text" json:"options"` // JSON array for select options
	Weight      int    `gorm:"column:weight;not null;default:0" json:"weight"`
	Enabled     bool   `gorm:"column:enabled;type:tinyint(2);not null;default:true" json:"enabled"`
	Created     int    `gorm:"column:created;not null" json:"created"` // Unix timestamp
	Updated     int    `gorm:"column:updated;not null" json:"updated"` // Unix timestamp
}

// TableName specifies the table name for the Catalog model
func (Catalog) TableName() string {
	return "ow_catalog"
}
