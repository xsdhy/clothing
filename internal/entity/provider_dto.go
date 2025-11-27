package entity

import (
	"strings"
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
	Modalities         []string               `json:"modalities"`
	SupportedSizes     []string               `json:"supported_sizes"`
	SupportedDurations []int                  `json:"supported_durations"`
	DefaultSize        string                 `json:"default_size"`
	DefaultDuration    *int                   `json:"default_duration"`
	Settings           map[string]interface{} `json:"settings"`
	IsActive           *bool                  `json:"is_active"`
}

// UpdateModelRequest defines payload for updating provider models.
type UpdateModelRequest struct {
	Name               *string                `json:"name"`
	Description        *string                `json:"description"`
	Price              *string                `json:"price"`
	MaxImages          *int                   `json:"max_images"`
	Modalities         *[]string              `json:"modalities"`
	SupportedSizes     *[]string              `json:"supported_sizes"`
	SupportedDurations *[]int                 `json:"supported_durations"`
	DefaultSize        *string                `json:"default_size"`
	DefaultDuration    *int                   `json:"default_duration"`
	Settings           map[string]interface{} `json:"settings"`
	IsActive           *bool                  `json:"is_active"`
}

// ProviderAdminView is the admin-facing provider representation.
type ProviderAdminView struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Driver      string                 `json:"driver"`
	Description string                 `json:"description,omitempty"`
	BaseURL     string                 `json:"base_url,omitempty"`
	Config      JSONMap                `json:"config,omitempty"`
	IsActive    bool                   `json:"is_active"`
	HasAPIKey   bool                   `json:"has_api_key"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Models      []ProviderModelSummary `json:"models,omitempty"`
}

// ProviderModelSummary is the admin-facing model representation.
type ProviderModelSummary struct {
	ModelID            string    `json:"model_id"`
	Name               string    `json:"name"`
	Description        string    `json:"description,omitempty"`
	Price              string    `json:"price,omitempty"`
	MaxImages          int       `json:"max_images,omitempty"`
	Modalities         []string  `json:"modalities,omitempty"`
	SupportedSizes     []string  `json:"supported_sizes,omitempty"`
	SupportedDurations []int     `json:"supported_durations,omitempty"`
	DefaultSize        string    `json:"default_size,omitempty"`
	DefaultDuration    int       `json:"default_duration,omitempty"`
	Settings           JSONMap   `json:"settings,omitempty"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToAdminView converts DbProvider to ProviderAdminView.
func (p DbProvider) ToAdminView(includeModels bool) ProviderAdminView {
	hasAPIKey := strings.TrimSpace(p.APIKey) != ""
	view := ProviderAdminView{
		ID:        p.ID,
		Name:      p.Name,
		Driver:    p.Driver,
		BaseURL:   p.BaseURL,
		Config:    p.Config,
		IsActive:  p.IsActive,
		HasAPIKey: hasAPIKey,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if strings.TrimSpace(p.Description) != "" {
		view.Description = p.Description
	}

	if includeModels {
		view.Models = make([]ProviderModelSummary, 0, len(p.Models))
		for _, model := range p.Models {
			view.Models = append(view.Models, model.ToAdminView())
		}
	}
	return view
}

// ToAdminView converts DbModel to ProviderModelSummary.
func (m DbModel) ToAdminView() ProviderModelSummary {
	modalities := []string(m.Modalities.ToSlice())
	sizes := []string(m.SupportedSizes.ToSlice())
	durations := mergeDurations(m.SupportedDurations, parseDurations(m.Settings))
	defaultSize := resolveDefaultSize(m.DefaultSize, sizes, m.Settings)
	defaultDuration := resolveDefaultDuration(m.DefaultDuration, durations, m.Settings)

	return ProviderModelSummary{
		ModelID:            m.ModelID,
		Name:               m.Name,
		Description:        m.Description,
		Price:              m.Price,
		MaxImages:          m.MaxImages,
		Modalities:         modalities,
		SupportedSizes:     sizes,
		SupportedDurations: durations,
		DefaultSize:        defaultSize,
		DefaultDuration:    defaultDuration,
		Settings:           m.Settings,
		IsActive:           m.IsActive,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}
