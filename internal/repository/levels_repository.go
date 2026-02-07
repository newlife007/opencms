package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// levelsRepository implements LevelsRepository
type levelsRepository struct {
	db *gorm.DB
}

// NewLevelsRepository creates a new levels repository
func NewLevelsRepository(db *gorm.DB) LevelsRepository {
	return &levelsRepository{db: db}
}

func (r *levelsRepository) Create(ctx context.Context, level *models.Levels) error {
	return r.db.WithContext(ctx).Create(level).Error
}

func (r *levelsRepository) FindByID(ctx context.Context, id int) (*models.Levels, error) {
	var level models.Levels
	err := r.db.WithContext(ctx).First(&level, id).Error
	if err != nil {
		return nil, err
	}
	return &level, nil
}

func (r *levelsRepository) FindAll(ctx context.Context) ([]*models.Levels, error) {
	var levels []*models.Levels
	// Order by level (ascending), then by id
	err := r.db.WithContext(ctx).Order("level ASC, id ASC").Find(&levels).Error
	return levels, err
}

func (r *levelsRepository) Update(ctx context.Context, level *models.Levels) error {
	return r.db.WithContext(ctx).Save(level).Error
}

func (r *levelsRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Levels{}, id).Error
}

// Levels returns levels repository implementation
func Levels(db *gorm.DB) LevelsRepository {
	return NewLevelsRepository(db)
}
