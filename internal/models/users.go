package models

// Users represents the ow_users table for user authentication and profile management
type Users struct {
	ID           int     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID      int     `gorm:"column:group_id;not null;index" json:"group_id"`
	LevelID      int     `gorm:"column:level_id;not null" json:"level_id"`
	Username     string  `gorm:"column:username;type:varchar(32);not null;uniqueIndex" json:"username"`
	Password     string  `gorm:"column:password;type:varchar(64);not null;index" json:"-"` // Never expose in JSON
	Nickname     string  `gorm:"column:nickname;type:varchar(64);not null" json:"nickname"`
	Sex          int     `gorm:"column:sex;type:tinyint(2);not null;default:0" json:"sex"` // 0:secret 1:male 2:female
	Birthday     *string `gorm:"column:birthday;type:varchar(64)" json:"birthday,omitempty"`
	Address      *string `gorm:"column:address;type:varchar(255)" json:"address,omitempty"`
	Email        *string `gorm:"column:email;type:varchar(64)" json:"email,omitempty"`
	Duty         *string `gorm:"column:duty;type:varchar(64)" json:"duty,omitempty"`
	OfficePhone  *string `gorm:"column:office_phone;type:varchar(64)" json:"office_phone,omitempty"`
	HomePhone    *string `gorm:"column:home_phone;type:varchar(64)" json:"home_phone,omitempty"`
	MobilePhone  *string `gorm:"column:mobile_phone;type:varchar(64)" json:"mobile_phone,omitempty"`
	Description  *string `gorm:"column:description;type:varchar(255)" json:"description,omitempty"`
	Enabled      bool    `gorm:"column:enabled;type:tinyint(2);not null;default:true" json:"enabled"`
	RegisterAt   int     `gorm:"column:register_at;not null;default:0" json:"register_at"` // Unix timestamp
	RegisterIP   string  `gorm:"column:register_ip;type:char(15);not null;default:'0.0.0.0'" json:"register_ip"`
	LoginCount   int     `gorm:"column:login_count;not null;default:0" json:"login_count"`
	LoginAt      int     `gorm:"column:login_at;not null;default:0" json:"login_at"` // Unix timestamp
	LoginIP      string  `gorm:"column:login_ip;type:char(15);not null;default:'0.0.0.0'" json:"login_ip"`

	// Relationships (not stored in database, populated via joins)
	Group *Groups `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Level *Levels `gorm:"foreignKey:LevelID" json:"level,omitempty"`
}

// TableName specifies the table name for the Users model
func (Users) TableName() string {
	return "ow_users"
}

// UserSex constants
const (
	UserSexSecret = 0
	UserSexMale   = 1
	UserSexFemale = 2
)
