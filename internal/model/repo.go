package model

import (
	"clothing/internal/entity"
	"context"
)

// Repository defines the interface for database operations
type Repository interface {

	// Usage records
	CreateUsageRecord(ctx context.Context, record *entity.DbUsageRecord) error
	ListUsageRecords(ctx context.Context, params *entity.UsageRecordQuery) ([]entity.DbUsageRecord, *entity.Meta, error)
	GetUsageRecord(ctx context.Context, id uint) (*entity.DbUsageRecord, error)
	DeleteUsageRecord(ctx context.Context, id uint) error
}
