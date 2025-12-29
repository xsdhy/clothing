package model

import (
	"clothing/internal/entity"
	"context"
)

// Repository 定义数据库操作接口
type Repository interface {
	// 用户管理
	CreateUser(ctx context.Context, user *entity.DbUser) error
	UpdateUser(ctx context.Context, id uint, updates entity.UserUpdates) error
	GetUserByEmail(ctx context.Context, email string) (*entity.DbUser, error)
	GetUserByID(ctx context.Context, id uint) (*entity.DbUser, error)
	ListUsers(ctx context.Context, params *entity.UserQuery) ([]entity.DbUser, *entity.Meta, error)
	DeleteUser(ctx context.Context, id uint) error
	CountUsers(ctx context.Context) (int64, error)

	// 使用记录
	CreateUsageRecord(ctx context.Context, record *entity.DbUsageRecord) error
	UpdateUsageRecord(ctx context.Context, id uint, updates entity.UsageRecordUpdates) error
	ListUsageRecords(ctx context.Context, params *entity.UsageRecordQuery) ([]entity.DbUsageRecord, *entity.Meta, error)
	GetUsageRecord(ctx context.Context, id uint) (*entity.DbUsageRecord, error)
	DeleteUsageRecord(ctx context.Context, id uint) error
	SetUsageRecordTags(ctx context.Context, recordID uint, tagIDs []uint) error
	ListTags(ctx context.Context) ([]entity.DbTag, error)
	CreateTag(ctx context.Context, tag *entity.DbTag) error
	UpdateTag(ctx context.Context, id uint, updates entity.TagUpdates) error
	DeleteTag(ctx context.Context, id uint) error
	FindTagsByIDs(ctx context.Context, ids []uint) ([]entity.DbTag, error)

	// 服务商和模型
	CreateProvider(ctx context.Context, provider *entity.DbProvider) error
	UpdateProvider(ctx context.Context, id string, updates entity.ProviderUpdates) error
	DeleteProvider(ctx context.Context, id string) error
	ListProviders(ctx context.Context, includeInactive bool) ([]entity.DbProvider, error)
	GetProvider(ctx context.Context, id string) (*entity.DbProvider, error)
	GetProviderWithModel(ctx context.Context, providerID, modelID string, includeInactive bool) (*entity.DbProvider, *entity.DbModel, error)

	GetModel(ctx context.Context, providerID, modelID string) (*entity.DbModel, error)
	CreateModel(ctx context.Context, model *entity.DbModel) error
	UpdateModel(ctx context.Context, providerID, modelID string, updates entity.ModelUpdates) error
	DeleteModel(ctx context.Context, providerID, modelID string) error
	ListModels(ctx context.Context, providerID string, includeInactive bool) ([]entity.DbModel, error)
}
