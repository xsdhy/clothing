package dto

import (
	"clothing/internal/entity/common"
	"time"
)

// CreateProviderRequest defines payload for creating providers.
type CreateProviderRequest struct {
	ID          string                 `json:"id" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Driver      string                 `json:"driver" binding:"required"`
	Description string                 `json:"description"`
	APIKey      string                 `json:"api_key"`
	BaseURL     string                 `json:"base_url"`
	Config      map[string]interface{} `json:"config"`
	IsActive    *bool                  `json:"is_active"`
}

// UpdateProviderRequest defines payload for updating providers.
type UpdateProviderRequest struct {
	Name        *string                `json:"name"`
	Driver      *string                `json:"driver"`
	Description *string                `json:"description"`
	APIKey      *string                `json:"api_key"`
	BaseURL     *string                `json:"base_url"`
	Config      map[string]interface{} `json:"config"`
	IsActive    *bool                  `json:"is_active"`
}

// CreateModelRequest defines payload for creating provider models.
type CreateModelRequest struct {
	ModelID            string                 `json:"model_id" binding:"required"`
	Name               string                 `json:"name" binding:"required"`
	Description        string                 `json:"description"`
	Price              string                 `json:"price"`
	MaxImages          *int                   `json:"max_images"`
	InputModalities    []string               `json:"input_modalities"`
	OutputModalities   []string               `json:"output_modalities"`
	SupportedSizes     []string               `json:"supported_sizes"`
	SupportedDurations []int                  `json:"supported_durations"`
	DefaultSize        string                 `json:"default_size"`
	DefaultDuration    *int                   `json:"default_duration"`
	Settings           map[string]interface{} `json:"settings"`
	// New fields for model capabilities
	GenerationMode    string `json:"generation_mode"`
	EndpointPath      string `json:"endpoint_path"`
	SupportsStreaming *bool  `json:"supports_streaming"`
	SupportsCancel    *bool  `json:"supports_cancel"`
	IsActive          *bool  `json:"is_active"`
}

// UpdateModelRequest defines payload for updating provider models.
type UpdateModelRequest struct {
	Name               *string                `json:"name"`
	Description        *string                `json:"description"`
	Price              *string                `json:"price"`
	MaxImages          *int                   `json:"max_images"`
	InputModalities    *[]string              `json:"input_modalities"`
	OutputModalities   *[]string              `json:"output_modalities"`
	SupportedSizes     *[]string              `json:"supported_sizes"`
	SupportedDurations *[]int                 `json:"supported_durations"`
	DefaultSize        *string                `json:"default_size"`
	DefaultDuration    *int                   `json:"default_duration"`
	Settings           map[string]interface{} `json:"settings"`
	// New fields for model capabilities
	GenerationMode    *string `json:"generation_mode"`
	EndpointPath      *string `json:"endpoint_path"`
	SupportsStreaming *bool   `json:"supports_streaming"`
	SupportsCancel    *bool   `json:"supports_cancel"`
	IsActive          *bool   `json:"is_active"`
}

// ProviderAdminView is the admin-facing provider representation.
type ProviderAdminView struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Driver      string                 `json:"driver"`
	Description string                 `json:"description,omitempty"`
	BaseURL     string                 `json:"base_url,omitempty"`
	Config      common.JSONMap         `json:"config,omitempty"`
	IsActive    bool                   `json:"is_active"`
	HasAPIKey   bool                   `json:"has_api_key"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Models      []ProviderModelSummary `json:"models,omitempty"`
}

// ProviderModelSummary is the admin-facing model representation.
type ProviderModelSummary struct {
	ModelID            string         `json:"model_id"`
	Name               string         `json:"name"`
	Description        string         `json:"description,omitempty"`
	Price              string         `json:"price,omitempty"`
	MaxImages          int            `json:"max_images,omitempty"`
	InputModalities    []string       `json:"input_modalities,omitempty"`
	OutputModalities   []string       `json:"output_modalities,omitempty"`
	SupportedSizes     []string       `json:"supported_sizes,omitempty"`
	SupportedDurations []int          `json:"supported_durations,omitempty"`
	DefaultSize        string         `json:"default_size,omitempty"`
	DefaultDuration    int            `json:"default_duration,omitempty"`
	Settings           common.JSONMap `json:"settings,omitempty"`
	// New fields for model capabilities
	GenerationMode    string `json:"generation_mode,omitempty"`
	EndpointPath      string `json:"endpoint_path,omitempty"`
	SupportsStreaming bool   `json:"supports_streaming"`
	SupportsCancel    bool   `json:"supports_cancel"`
	IsActive          bool   `json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}
