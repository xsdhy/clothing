package entity

// Re-export DB entities from the db package for backward compatibility.

import (
	"clothing/internal/entity/db"
)

// Type aliases for DB entities
type DbUser = db.User
type DbProvider = db.Provider
type DbModel = db.Model
type DbUsageRecord = db.UsageRecord
type DbTag = db.Tag
type DbUsageRecordTag = db.UsageRecordTag

// User role constants
const (
	UserRoleSuperAdmin = db.UserRoleSuperAdmin
	UserRoleAdmin      = db.UserRoleAdmin
	UserRoleUser       = db.UserRoleUser
)

// Provider driver constants
const (
	ProviderDriverOpenRouter = db.ProviderDriverOpenRouter
	ProviderDriverGemini     = db.ProviderDriverGemini
	ProviderDriverAiHubMix   = db.ProviderDriverAiHubMix
	ProviderDriverDashscope  = db.ProviderDriverDashscope
	ProviderDriverFal        = db.ProviderDriverFal
	ProviderDriverVolcengine = db.ProviderDriverVolcengine
)
