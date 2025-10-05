package sql

import (
	"clothing/internal/entity"
	"context"
	"fmt"
	"strings"
)

// CreateUsageRecord inserts a new usage record into the database.
func (r *GormRepository) CreateUsageRecord(ctx context.Context, record *entity.DbUsageRecord) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if record == nil {
		return fmt.Errorf("record is nil")
	}
	return r.db.WithContext(ctx).Create(record).Error
}

// ListUsageRecords retrieves paginated usage records.
func (r *GormRepository) ListUsageRecords(ctx context.Context, params *entity.UsageRecordQuery) ([]entity.DbUsageRecord, *entity.Meta, error) {
	if r == nil || r.db == nil {
		return nil, nil, fmt.Errorf("repository not initialised")
	}

	query := r.db.WithContext(ctx).Model(&entity.DbUsageRecord{})
	if params != nil {
		if trimmed := strings.TrimSpace(params.Provider); trimmed != "" {
			query = query.Where("provider_id = ?", trimmed)
		}
		if trimmed := strings.TrimSpace(params.Model); trimmed != "" {
			query = query.Where("model_id = ?", trimmed)
		}
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}

	page := 1
	pageSize := 20
	if params != nil {
		if params.Page > 0 {
			page = int(params.Page)
		}
		if params.PageSize > 0 {
			pageSize = int(params.PageSize)
		}
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	var records []entity.DbUsageRecord
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, nil, err
	}

	meta := r.calculatePagination(totalCount, page, pageSize)
	return records, meta, nil
}
