package repository

import (
	"context"
	"strings"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// usersRepository implements UsersRepository
type usersRepository struct {
	db *gorm.DB
}

// NewUsersRepository creates a new users repository
func NewUsersRepository(db *gorm.DB) UsersRepository {
	return &usersRepository{db: db}
}

func (r *usersRepository) Create(ctx context.Context, user *models.Users) error {
	// IMPORTANT WORKAROUND for GORM zero value issue:
	// GORM has a known limitation where it ignores false values for bool fields during INSERT,
	// even when explicitly listed in Select(). The database default (true) takes precedence.
	// After Create(), GORM also updates the struct with the database default value.
	// 
	// Solution: Save the original enabled value, then explicitly UPDATE it if it was false.
	
	originalEnabled := user.Enabled
	
	// Create the user record (database will use default enabled=true)
	result := r.db.WithContext(ctx).Omit("ID", "Group", "Level").Create(user)
	if result.Error != nil {
		return result.Error
	}
	
	// If the original enabled value was false, explicitly update it
	// (UPDATE statements don't skip zero values like INSERT does)
	if !originalEnabled {
		updateResult := r.db.WithContext(ctx).Model(&models.Users{}).Where("id = ?", user.ID).Update("enabled", false)
		if updateResult.Error != nil {
			return updateResult.Error
		}
		// Update the struct to reflect the correct value
		user.Enabled = false
	}
	
	return nil
}

func (r *usersRepository) FindByID(ctx context.Context, id int) (*models.Users, error) {
	var user models.Users
	err := r.db.WithContext(ctx).Preload("Group").Preload("Level").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *usersRepository) FindByUsername(ctx context.Context, username string) (*models.Users, error) {
	var user models.Users
	err := r.db.WithContext(ctx).Preload("Group").Preload("Level").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *usersRepository) FindAll(ctx context.Context, limit, offset int) ([]*models.Users, int64, error) {
	var users []*models.Users
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.Users{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Preload("Group").Preload("Level").
		Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}

func (r *usersRepository) Update(ctx context.Context, user *models.Users) error {
	return r.db.WithContext(ctx).Model(&models.Users{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"password":  user.Password,
		"nickname":  user.Nickname,
		"email":     user.Email,
		"group_id":  user.GroupID,
		"level_id":  user.LevelID,
		"enabled":   user.Enabled,
	}).Error
}

func (r *usersRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Users{}, id).Error
}

func (r *usersRepository) UpdatePassword(ctx context.Context, id int, hashedPassword string) error {
	return r.db.WithContext(ctx).Model(&models.Users{}).
		Where("id = ?", id).
		Update("password", hashedPassword).Error
}

// UpdateStatus updates user enabled status
func (r *usersRepository) UpdateStatus(ctx context.Context, id uint, status int) error {
	return r.db.WithContext(ctx).Model(&models.Users{}).
		Where("id = ?", id).
		Update("enabled", status).Error
}

// GetByID gets user by ID with relationships (for service layer)
func (r *usersRepository) GetByID(ctx context.Context, id uint) (*models.Users, error) {
	var user models.Users
	err := r.db.WithContext(ctx).
		Preload("Group").
		Preload("Group.Roles").
		Preload("Group.Roles.Permissions").
		Preload("Level").
		First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *usersRepository) UpdateLoginInfo(ctx context.Context, id int, ip string) error {
	return r.db.WithContext(ctx).Model(&models.Users{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"login_count": gorm.Expr("login_count + 1"),
			"login_at":    gorm.Expr("UNIX_TIMESTAMP()"),
			"login_ip":    ip,
		}).Error
}

// UserSearchParams defines search parameters for users
type UserSearchParams struct {
	Username string
	Nickname string
	Email    string
	GroupIDs []int
	LevelIDs []int
	Enabled  *bool
	Page     int
	PageSize int
}

// SearchUsers searches users with filters
func (r *usersRepository) SearchUsers(ctx context.Context, params UserSearchParams) ([]*models.Users, int64, error) {
	var users []*models.Users
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Users{})

	// Apply filters
	if params.Username != "" {
		// Support wildcard search: * -> %
		username := strings.ReplaceAll(params.Username, "*", "%")
		if !strings.Contains(username, "%") {
			username = "%" + username + "%"
		}
		query = query.Where("username LIKE ?", username)
	}

	if params.Nickname != "" {
		nickname := strings.ReplaceAll(params.Nickname, "*", "%")
		if !strings.Contains(nickname, "%") {
			nickname = "%" + nickname + "%"
		}
		query = query.Where("nickname LIKE ?", nickname)
	}

	if params.Email != "" {
		email := strings.ReplaceAll(params.Email, "*", "%")
		if !strings.Contains(email, "%") {
			email = "%" + email + "%"
		}
		query = query.Where("email LIKE ?", email)
	}

	if len(params.GroupIDs) > 0 {
		query = query.Where("group_id IN ?", params.GroupIDs)
	}

	if len(params.LevelIDs) > 0 {
		query = query.Where("level_id IN ?", params.LevelIDs)
	}

	if params.Enabled != nil {
		query = query.Where("enabled = ?", *params.Enabled)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 12 // Default from PHP system
	}
	offset := (params.Page - 1) * params.PageSize

	// Fetch results with preloading
	err := query.Preload("Group").Preload("Level").
		Order("id DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// BatchDelete deletes multiple users
func (r *usersRepository) BatchDelete(ctx context.Context, ids []int) error {
	return r.db.WithContext(ctx).Delete(&models.Users{}, ids).Error
}

// CheckUsernameExists checks if username already exists
func (r *usersRepository) CheckUsernameExists(ctx context.Context, username string, excludeID int) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Users{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
