package sql

import (
	"clothing/internal/entity"
	"context"
)

// Demo 设备列表
func (r *GormRepository) GetDemoList(ctx context.Context, params *entity.BaseParams) ([]entity.DbDemo, *entity.Meta, error) {
	var records []entity.DbDemo
	var totalCount int64

	query := r.db.WithContext(ctx)

	//if params.DeviceStatus != "" {
	//	query = query.Where("device_status = ?", params.DeviceStatus)
	//}

	// Count total records
	err := query.Model(&entity.DbDemo{}).Count(&totalCount).Error
	if err != nil {
		return nil, nil, err
	}

	// Apply pagination
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(int(offset)).Limit(int(params.PageSize))
	}

	// Apply sorting
	if params.SortBy != "" {
		direction := "ASC"
		if params.SortDesc {
			direction = "DESC"
		}
		query = query.Order(params.SortBy + " " + direction)
	} else {
		// Default sort by creation time descending
		query = query.Order("id DESC")
	}

	// Execute query
	err = query.Find(&records).Error
	if err != nil {
		return nil, nil, err
	}

	// Calculate pagination metadata
	meta := r.calculatePagination(totalCount, int(params.Page), int(params.PageSize))

	return records, meta, nil
}
