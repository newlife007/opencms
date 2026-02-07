package repository

import (
	"testing"
	"time"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.Role{},
		&models.Permission{},
		&models.Level{},
		&models.Category{},
		&models.File{},
		&models.Catalog{},
	)
	assert.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
		GroupID:  1,
		LevelID:  1,
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	// Verify user was created
	found, err := repo.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, found.Username)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
		GroupID:  1,
		LevelID:  1,
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	// Get by username
	found, err := repo.GetByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)

	// Get non-existent user
	_, err = repo.GetByUsername("nonexistent")
	assert.Error(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
		GroupID:  1,
		LevelID:  1,
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	// Update email
	user.Email = "newemail@example.com"
	err = repo.Update(user)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "newemail@example.com", found.Email)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
		GroupID:  1,
		LevelID:  1,
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	// Delete user
	err = repo.Delete(user.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(user.ID)
	assert.Error(t, err)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create multiple users
	users := []*models.User{
		{Username: "user1", Password: "pass1", Email: "user1@example.com", GroupID: 1, LevelID: 1},
		{Username: "user2", Password: "pass2", Email: "user2@example.com", GroupID: 1, LevelID: 1},
		{Username: "user3", Password: "pass3", Email: "user3@example.com", GroupID: 2, LevelID: 2},
	}

	for _, user := range users {
		err := repo.Create(user)
		assert.NoError(t, err)
	}

	// List all users
	found, total, err := repo.List(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, found, 3)
}

func TestCategoryRepository_GetTree(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCategoryRepository(db)

	// Create category hierarchy
	root := &models.Category{
		Name:        "Root",
		Description: "Root category",
		ParentID:    0,
		Path:        "0",
		Weight:      1,
		Status:      1,
	}
	err := repo.Create(root)
	assert.NoError(t, err)

	child1 := &models.Category{
		Name:        "Child 1",
		Description: "First child",
		ParentID:    int64(root.ID),
		Path:        "0," + string(rune(root.ID)),
		Weight:      1,
		Status:      1,
	}
	err = repo.Create(child1)
	assert.NoError(t, err)

	child2 := &models.Category{
		Name:        "Child 2",
		Description: "Second child",
		ParentID:    int64(root.ID),
		Path:        "0," + string(rune(root.ID)),
		Weight:      2,
		Status:      1,
	}
	err = repo.Create(child2)
	assert.NoError(t, err)

	// Get tree
	tree, err := repo.GetTree()
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	// Tree should have root category with children
	assert.Greater(t, len(tree), 0)
}

func TestFileRepository_GetByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFileRepository(db)

	// Create files with different statuses
	files := []*models.File{
		{
			Title:          "File 1",
			Name:           "file1.mp4",
			Type:           1,
			Status:         0, // New
			Path:           "/path/file1.mp4",
			Size:           1024,
			UploadUsername: "user1",
		},
		{
			Title:          "File 2",
			Name:           "file2.mp4",
			Type:           1,
			Status:         2, // Published
			Path:           "/path/file2.mp4",
			Size:           2048,
			UploadUsername: "user1",
		},
		{
			Title:          "File 3",
			Name:           "file3.mp4",
			Type:           1,
			Status:         2, // Published
			Path:           "/path/file3.mp4",
			Size:           3072,
			UploadUsername: "user2",
		},
	}

	for _, file := range files {
		err := repo.Create(file)
		assert.NoError(t, err)
	}

	// Get published files
	filters := map[string]interface{}{
		"status": 2,
	}
	found, total, err := repo.List(1, 10, filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, found, 2)

	// Verify all returned files have status 2
	for _, file := range found {
		assert.Equal(t, 2, file.Status)
	}
}

func TestFileRepository_GetByType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFileRepository(db)

	// Create files with different types
	files := []*models.File{
		{Title: "Video", Name: "video.mp4", Type: 1, Status: 2, Path: "/path/video.mp4", Size: 1024, UploadUsername: "user1"},
		{Title: "Audio", Name: "audio.mp3", Type: 2, Status: 2, Path: "/path/audio.mp3", Size: 512, UploadUsername: "user1"},
		{Title: "Image", Name: "image.jpg", Type: 3, Status: 2, Path: "/path/image.jpg", Size: 256, UploadUsername: "user1"},
	}

	for _, file := range files {
		err := repo.Create(file)
		assert.NoError(t, err)
	}

	// Get video files (type 1)
	filters := map[string]interface{}{
		"type": 1,
	}
	found, total, err := repo.List(1, 10, filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, found, 1)
	assert.Equal(t, 1, found[0].Type)
}
