package service

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/repository"
)

// LevelsService handles business logic for levels
type LevelsService struct {
	repo repository.LevelsRepository
}

// NewLevelsService creates a new levels service
func NewLevelsService(repo repository.LevelsRepository) *LevelsService {
	return &LevelsService{
		repo: repo,
	}
}

// GetAll returns all levels
func (s *LevelsService) GetAll(ctx context.Context) ([]*models.Levels, error) {
	return s.repo.FindAll(ctx)
}

// GetByID returns a level by ID
func (s *LevelsService) GetByID(ctx context.Context, id int) (*models.Levels, error) {
	return s.repo.FindByID(ctx, id)
}

// Create creates a new level
func (s *LevelsService) Create(ctx context.Context, level *models.Levels) error {
	return s.repo.Create(ctx, level)
}

// Update updates a level
func (s *LevelsService) Update(ctx context.Context, level *models.Levels) error {
	return s.repo.Update(ctx, level)
}

// Delete deletes a level
func (s *LevelsService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
