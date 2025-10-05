package model

import (
	"clothing/internal/entity"
	"context"
)

// Repository defines the interface for database operations
type Repository interface {

	// Demo
	GetDemoList(ctx context.Context, params *entity.BaseParams) ([]entity.DbDemo, *entity.Meta, error)
}
