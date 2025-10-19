package model

import (
	"clothing/internal/entity"
	"context"
)

// Repository defines the interface for database operations
type Repository interface {
	// Users
	CreateUser(ctx context.Context, user *entity.DbUser) error
	UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) error
	GetUserByEmail(ctx context.Context, email string) (*entity.DbUser, error)
	GetUserByID(ctx context.Context, id uint) (*entity.DbUser, error)
	ListUsers(ctx context.Context, params *entity.UserQuery) ([]entity.DbUser, *entity.Meta, error)
	DeleteUser(ctx context.Context, id uint) error
	CountUsers(ctx context.Context) (int64, error)

	// Usage records
	CreateUsageRecord(ctx context.Context, record *entity.DbUsageRecord) error
	ListUsageRecords(ctx context.Context, params *entity.UsageRecordQuery) ([]entity.DbUsageRecord, *entity.Meta, error)
	GetUsageRecord(ctx context.Context, id uint) (*entity.DbUsageRecord, error)
	DeleteUsageRecord(ctx context.Context, id uint) error
}
