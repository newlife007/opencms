package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicateUsername = errors.New("username already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrWeakPassword      = errors.New("password must be 3-32 characters")
)

// UsersService handles business logic for users
type UsersService struct {
	usersRepo  repository.UsersRepository
	groupsRepo repository.GroupsRepository
	levelsRepo repository.LevelsRepository
}

// NewUsersService creates a new users service
func NewUsersService(
	usersRepo repository.UsersRepository,
	groupsRepo repository.GroupsRepository,
	levelsRepo repository.LevelsRepository,
) *UsersService {
	return &UsersService{
		usersRepo:  usersRepo,
		groupsRepo: groupsRepo,
		levelsRepo: levelsRepo,
	}
}

// UserListRequest represents request for listing users
type UserListRequest struct {
	Username string  `json:"username" form:"username"`
	Nickname string  `json:"nickname" form:"nickname"`
	Email    string  `json:"email" form:"email"`
	GroupIDs []int   `json:"group_ids" form:"group_ids"`
	LevelIDs []int   `json:"level_ids" form:"level_ids"`
	Enabled  *bool   `json:"enabled" form:"enabled"`
	Page     int     `json:"page" form:"page"`
	PageSize int     `json:"page_size" form:"page_size"`
}

// UserCreateRequest represents request for creating a user
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=3,max=32"`
	Nickname string `json:"nickname" binding:"required,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=255"`
	GroupID  int    `json:"group_id" binding:"required"`
	LevelID  int    `json:"level_id" binding:"required"`
	Enabled  bool   `json:"enabled"`
}

// UserUpdateRequest represents request for updating a user
type UserUpdateRequest struct {
	ID       int    `json:"id"` // Set from URL param, not from JSON
	Password string `json:"password" binding:"omitempty,min=3,max=32"` // Optional
	Nickname string `json:"nickname" binding:"required,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=255"`
	GroupID  int    `json:"group_id" binding:"required"`
	LevelID  int    `json:"level_id" binding:"required"`
	Enabled  bool   `json:"enabled"`
}

// UserResponse represents user response
type UserResponse struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	GroupID    int    `json:"group_id"`
	GroupName  string `json:"group_name"`
	LevelID    int    `json:"level_id"`
	LevelName  string `json:"level_name"`
	Enabled    bool   `json:"enabled"`
	LoginCount int    `json:"login_count"`
	LoginAt    int    `json:"login_at"`
	LoginIP    string `json:"login_ip"`
	RegisterAt int    `json:"register_at"`
	RegisterIP string `json:"register_ip"`
}

// ListUsers lists users with filters
func (s *UsersService) ListUsers(ctx context.Context, req UserListRequest) ([]*UserResponse, int64, error) {
	params := repository.UserSearchParams{
		Username: req.Username,
		Nickname: req.Nickname,
		Email:    req.Email,
		GroupIDs: req.GroupIDs,
		LevelIDs: req.LevelIDs,
		Enabled:  req.Enabled,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	users, total, err := s.usersRepo.SearchUsers(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(user)
	}

	return responses, total, nil
}

// GetUser gets a user by ID
func (s *UsersService) GetUser(ctx context.Context, id int) (*UserResponse, error) {
	user, err := s.usersRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return s.toUserResponse(user), nil
}

// CreateUser creates a new user
func (s *UsersService) CreateUser(ctx context.Context, req UserCreateRequest) (*UserResponse, error) {
	// Validate password length
	if len(req.Password) < 3 || len(req.Password) > 32 {
		return nil, ErrWeakPassword
	}

	// Check if username exists
	exists, err := s.usersRepo.CheckUsernameExists(ctx, req.Username, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateUsername
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	
	// Convert email to pointer
	var emailPtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}
	
	user := &models.Users{
		Username:   req.Username,
		Password:   string(hashedPassword),
		Nickname:   req.Nickname,
		Email:      emailPtr,
		GroupID:    req.GroupID,
		LevelID:    req.LevelID,
		Enabled:    req.Enabled,
		LoginCount: 0,
		RegisterAt: int(now),
		RegisterIP: "0.0.0.0", // TODO: get from request context
	}

	if err := s.usersRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Reload user with associations
	user, err = s.usersRepo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

// UpdateUser updates a user
func (s *UsersService) UpdateUser(ctx context.Context, req UserUpdateRequest) (*UserResponse, error) {
	// Get existing user
	user, err := s.usersRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update password if provided
	if req.Password != "" {
		if len(req.Password) < 3 || len(req.Password) > 32 {
			return nil, ErrWeakPassword
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	// Update other fields
	user.Nickname = req.Nickname
	
	// Convert email to pointer
	if req.Email != "" {
		user.Email = &req.Email
	} else {
		user.Email = nil
	}
	
	user.GroupID = req.GroupID
	user.LevelID = req.LevelID
	user.Enabled = req.Enabled
	
	// Clear preloaded associations to avoid GORM using their IDs
	user.Group = nil
	user.Level = nil

	if err := s.usersRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Reload user with associations
	user, err = s.usersRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

// DeleteUser deletes a user
func (s *UsersService) DeleteUser(ctx context.Context, id int) error {
	// Check if user exists
	_, err := s.usersRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.usersRepo.Delete(ctx, id)
}

// BatchDeleteUsers deletes multiple users
func (s *UsersService) BatchDeleteUsers(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return errors.New("no user IDs provided")
	}
	return s.usersRepo.BatchDelete(ctx, ids)
}

// Helper function to convert model to response
func (s *UsersService) toUserResponse(user *models.Users) *UserResponse {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	
	response := &UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Email:      email,
		GroupID:    user.GroupID,
		LevelID:    user.LevelID,
		Enabled:    user.Enabled,
		LoginCount: user.LoginCount,
		LoginAt:    user.LoginAt,
		LoginIP:    user.LoginIP,
		RegisterAt: user.RegisterAt,
		RegisterIP: user.RegisterIP,
	}

	if user.Group != nil {
		response.GroupName = user.Group.Name
	}

	if user.Level != nil {
		response.LevelName = user.Level.Name
	}

	return response
}


// ResetPassword resets a user's password (admin operation)
func (s *UsersService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	// Validate password
	if len(newPassword) < 6 || len(newPassword) > 32 {
		return errors.New("password must be 6-32 characters")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password in database (convert uint to int for repository)
	if err := s.usersRepo.UpdatePassword(ctx, int(userID), string(hashedPassword)); err != nil {
		return err
	}

	// IMPORTANT: Enable the account when resetting password
	// This ensures the user can login with the new password
	return s.usersRepo.UpdateStatus(ctx, userID, 1)
}

// UpdateUserStatus updates a user's enabled status
func (s *UsersService) UpdateUserStatus(ctx context.Context, userID uint, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	
	return s.usersRepo.UpdateStatus(ctx, userID, enabledInt)
}

// GetUserPermissions gets a user's effective permissions
func (s *UsersService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	// Get user with relationships
	user, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Collect permissions from user's group roles
	permissionSet := make(map[string]bool)
	
	if user.Group != nil && user.Group.Roles != nil {
		for _, role := range user.Group.Roles {
			if role.Permissions != nil {
				for _, perm := range role.Permissions {
					// Build permission key from namespace.controller.action
					permKey := fmt.Sprintf("%s.%s.%s", perm.Namespace, perm.Controller, perm.Action)
					permissionSet[permKey] = true
				}
			}
		}
	}

	// Convert set to slice
	permissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

