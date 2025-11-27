package sql

import (
	"clothing/internal/entity"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
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

// UpdateUsageRecord updates a usage record with the provided fields.
func (r *GormRepository) UpdateUsageRecord(ctx context.Context, id uint, updates map[string]interface{}) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return fmt.Errorf("invalid usage record id")
	}
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}
	return r.db.WithContext(ctx).Model(&entity.DbUsageRecord{}).Where("id = ?", id).Updates(updates).Error
}

// ListUsageRecords retrieves paginated usage records.
func (r *GormRepository) ListUsageRecords(ctx context.Context, params *entity.UsageRecordQuery) ([]entity.DbUsageRecord, *entity.Meta, error) {
	if r == nil || r.db == nil {
		return nil, nil, fmt.Errorf("repository not initialised")
	}

	query := r.db.WithContext(ctx).
		Model(&entity.DbUsageRecord{}).
		Preload("User").
		Preload("Tags")
	if params != nil {
		if trimmed := strings.TrimSpace(params.Provider); trimmed != "" {
			query = query.Where("provider_id = ?", trimmed)
		}
		if trimmed := strings.TrimSpace(params.Model); trimmed != "" {
			query = query.Where("model_id = ?", trimmed)
		}
		if !params.IncludeAll && params.UserID > 0 {
			query = query.Where("user_id = ?", params.UserID)
		}
		if trimmed := strings.ToLower(strings.TrimSpace(params.Result)); trimmed != "" && trimmed != "all" {
			switch trimmed {
			case "success":
				query = query.Where("(error_message IS NULL OR error_message = '')")
			case "failure":
				query = query.Where("error_message IS NOT NULL AND error_message <> ''")
			}
		}

		if len(params.TagIDs) > 0 {
			query = query.Joins("JOIN usage_record_tags ON usage_record_tags.usage_record_id = usage_records.id").
				Where("usage_record_tags.tag_id IN ?", params.TagIDs).
				Group("usage_records.id").
				Having("COUNT(DISTINCT usage_record_tags.tag_id) >= ?", len(params.TagIDs))
		}
		if params.HasOutputImages {
			query = r.applyHasOutputImagesFilter(query)
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
	if err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, nil, err
	}

	meta := r.calculatePagination(totalCount, page, pageSize)
	return records, meta, nil
}

func (r *GormRepository) applyHasOutputImagesFilter(query *gorm.DB) *gorm.DB {
	if query == nil {
		return query
	}

	dialect := ""
	if r != nil && r.db != nil && r.db.Dialector != nil {
		dialect = strings.ToLower(r.db.Dialector.Name())
	}

	switch dialect {
	case "mysql":
		return query.Where("JSON_LENGTH(output_images) > 0")
	case "sqlite":
		return query.Where("json_array_length(output_images) > 0")
	case "postgres":
		return query.Where("json_array_length(output_images::json) > 0")
	default:
		return query.Where("output_images IS NOT NULL AND output_images <> '' AND output_images <> '[]'")
	}
}

// GetUsageRecord retrieves a single usage record by ID.
func (r *GormRepository) GetUsageRecord(ctx context.Context, id uint) (*entity.DbUsageRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return nil, fmt.Errorf("invalid usage record id")
	}

	var record entity.DbUsageRecord
	if err := r.db.WithContext(ctx).Preload("User").Preload("Tags").First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to load usage record: %w", err)
	}
	return &record, nil
}

// DeleteUsageRecord removes a usage record by ID.
func (r *GormRepository) DeleteUsageRecord(ctx context.Context, id uint) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return fmt.Errorf("invalid usage record id")
	}

	result := r.db.WithContext(ctx).Delete(&entity.DbUsageRecord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
