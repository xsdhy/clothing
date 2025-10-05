package model

import (
	"clothing/internal/entity"
	"context"
)

// Repository defines the interface for database operations
type Repository interface {

	// Demo
	GetDemoList(ctx context.Context, params *entity.BaseParams) ([]entity.DbDemo, *entity.Meta, error)

	// Usage records
	CreateUsageRecord(ctx context.Context, record *entity.DbUsageRecord) error
	ListUsageRecords(ctx context.Context, params *entity.UsageRecordQuery) ([]entity.DbUsageRecord, *entity.Meta, error)
}
