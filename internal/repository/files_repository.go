package repository

import (
	"context"

	"github.com/openwan/media-asset-management/internal/models"
	"gorm.io/gorm"
)

// filesRepository implements FilesRepository
type filesRepository struct {
	db *gorm.DB
}

// NewFilesRepository creates a new files repository
func NewFilesRepository(db *gorm.DB) FilesRepository {
	return &filesRepository{db: db}
}

func (r *filesRepository) Create(ctx context.Context, file *models.Files) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *filesRepository) FindByID(ctx context.Context, id uint64) (*models.Files, error) {
	var file models.Files
	err := r.db.WithContext(ctx).First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *filesRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Files, int64, error) {
	var files []*models.Files
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Files{})

	// Apply filters
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if fileType, ok := filters["type"]; ok {
		query = query.Where("type = ?", fileType)
	}
	if categoryID, ok := filters["category_id"]; ok {
		query = query.Where("category_id = ?", categoryID)
	}
	if level, ok := filters["level"]; ok {
		query = query.Where("level <= ?", level)
	}
	if uploadDateFrom, ok := filters["upload_date_from"]; ok {
		query = query.Where("upload_at >= ?", uploadDateFrom)
	}
	if uploadDateTo, ok := filters["upload_date_to"]; ok {
		query = query.Where("upload_at <= ?", uploadDateTo)
	}
	
	// Add text search filter (CRITICAL FIX: This was missing!)
	if searchQuery, ok := filters["search_query"]; ok {
		queryStr := searchQuery.(string)
		// Search in title, name, and catalog_info JSON fields
		query = query.Where(
			"title LIKE ? OR name LIKE ? OR ext LIKE ? OR catalog_info LIKE ?",
			"%"+queryStr+"%",
			"%"+queryStr+"%",
			"%"+queryStr+"%",
			"%"+queryStr+"%",
		)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("upload_at DESC").Limit(limit).Offset(offset).Find(&files).Error
	return files, total, err
}

func (r *filesRepository) Update(ctx context.Context, file *models.Files) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *filesRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Files{}, id).Error
}

func (r *filesRepository) FindByMD5(ctx context.Context, md5 string) (*models.Files, error) {
	var file models.Files
	err := r.db.WithContext(ctx).Where("name = ?", md5).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

func (r *filesRepository) FindByStatusAndType(ctx context.Context, status, fileType int, limit, offset int) ([]*models.Files, int64, error) {
	var files []*models.Files
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Files{}).Where("status = ? AND type = ?", status, fileType)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("upload_at DESC").Limit(limit).Offset(offset).Find(&files).Error
	return files, total, err
}

func (r *filesRepository) UpdateStatus(ctx context.Context, id uint64, status int, username string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case models.FileStatusPending:
		updates["catalog_username"] = username
		updates["catalog_at"] = gorm.Expr("UNIX_TIMESTAMP()")
	case models.FileStatusPublished:
		updates["putout_username"] = username
		updates["putout_at"] = gorm.Expr("UNIX_TIMESTAMP()")
	case models.FileStatusRejected:
		updates["putout_username"] = username
		updates["putout_at"] = gorm.Expr("UNIX_TIMESTAMP()")
	}

	return r.db.WithContext(ctx).Model(&models.Files{}).Where("id = ?", id).Updates(updates).Error
}
