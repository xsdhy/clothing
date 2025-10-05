package sql

import (
	"clothing/internal/entity"

	"gorm.io/gorm"
)

// GormRepository implements Repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new repository instance
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// calculatePagination calculates pagination metrics
func (r *GormRepository) calculatePagination(totalCount int64, page, pageSize int) *dto.Meta {
	meta := &entity.Meta{
		Total:    totalCount,
		PageSize: int64(pageSize),
	}

	return meta
}
