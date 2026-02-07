package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/repository"
	"github.com/openwan/media-asset-management/pkg/crypto"
)

// ACLService handles ACL-related business logic
type ACLService struct {
	repo repository.Repository
}

// NewACLService creates a new ACL service
func NewACLService(repo repository.Repository) *ACLService {
	return &ACLService{repo: repo}
}

// CreateUser creates a new user with hashed password
func (s *ACLService) CreateUser(ctx context.Context, user *models.Users, password string) error {
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword
	return s.repo.Users().Create(ctx, user)
}

// GetUser retrieves a user by ID
func (s *ACLService) GetUser(ctx context.Context, id int) (*models.Users, error) {
	return s.repo.Users().FindByID(ctx, id)
}

// GetUserByID retrieves a user by ID (uint version for handlers)
func (s *ACLService) GetUserByID(ctx context.Context, id uint) (*models.Users, error) {
	return s.repo.Users().FindByID(ctx, int(id))
}

// ListUsers retrieves all users with pagination
func (s *ACLService) ListUsers(ctx context.Context, limit, offset int) ([]*models.Users, int64, error) {
	return s.repo.Users().FindAll(ctx, limit, offset)
}

// UpdateUser updates user information
func (s *ACLService) UpdateUser(ctx context.Context, user *models.Users) error {
	return s.repo.Users().Update(ctx, user)
}

// DeleteUser deletes a user (with validation)
func (s *ACLService) DeleteUser(ctx context.Context, id int) error {
	// TODO: Add validation to prevent deleting users with active files
	return s.repo.Users().Delete(ctx, id)
}

// ResetPassword resets a user's password
func (s *ACLService) ResetPassword(ctx context.Context, userID int, newPassword string) error {
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.Users().UpdatePassword(ctx, userID, hashedPassword)
}

// AuthenticateUser authenticates a user by username and password
func (s *ACLService) AuthenticateUser(ctx context.Context, username, password string) (*models.Users, error) {
	user, err := s.repo.Users().FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !crypto.CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	if !user.Enabled {
		return nil, fmt.Errorf("user account is disabled")
	}

	return user, nil
}

// UpdateLoginInfo updates user's login information
func (s *ACLService) UpdateLoginInfo(ctx context.Context, userID int, ip string) error {
	return s.repo.Users().UpdateLoginInfo(ctx, userID, ip)
}

// HasPermission checks if user has a specific permission
func (s *ACLService) HasPermission(ctx context.Context, userID int, namespace, controller, action string) (bool, error) {
	return s.repo.ACL().HasPermission(ctx, userID, namespace, controller, action)
}

// GetUserPermissions retrieves all permissions for a user
func (s *ACLService) GetUserPermissions(ctx context.Context, userID int) ([]*models.Permissions, error) {
	return s.repo.ACL().GetUserPermissions(ctx, userID)
}

// GetUserRoles retrieves all roles assigned to a user through their group
func (s *ACLService) GetUserRoles(ctx context.Context, userID int) ([]*models.Roles, error) {
	return s.repo.ACL().GetUserRoles(ctx, userID)
}

// IsAdmin checks if user is an administrator
func (s *ACLService) IsAdmin(ctx context.Context, userID int) (bool, error) {
	return s.repo.ACL().IsAdmin(ctx, userID)
}

// CanAccessCategory checks if user can access a category
func (s *ACLService) CanAccessCategory(ctx context.Context, userID, categoryID int) (bool, error) {
	return s.repo.ACL().CanAccessCategory(ctx, userID, categoryID)
}

// CanAccessFile checks if user can access a file
func (s *ACLService) CanAccessFile(ctx context.Context, userID int, fileID uint64) (bool, error) {
	return s.repo.ACL().CanAccessFile(ctx, userID, fileID)
}


// UpdateUserProfile updates user profile information
func (s *ACLService) UpdateUserProfile(ctx context.Context, userID int, email, realName, telephone string) error {
	// Get user first
	user, err := s.repo.Users().FindByID(ctx, userID)
	if err != nil {
		return err
	}
	
	if user == nil {
		return errors.New("user not found")
	}

	// Update fields (use pointers for nullable fields)
	if email != "" {
		user.Email = &email
	}
	// Note: Users model doesn't have RealName and Telephone as separate fields
	// Using existing fields from the schema (duty for real name, mobile_phone for telephone)
	if realName != "" {
		user.Duty = &realName // Using duty field to store real name
	}
	if telephone != "" {
		user.MobilePhone = &telephone
	}

	// Save changes
	return s.repo.Users().Update(ctx, user)
}

