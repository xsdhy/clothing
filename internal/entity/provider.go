package entity

import (
	"strings"
	"time"
)

const (
	ProviderDriverOpenRouter = "openrouter"
	ProviderDriverGemini     = "gemini"
	ProviderDriverAiHubMix   = "aihubmix"
	ProviderDriverDashscope  = "dashscope"
	ProviderDriverFal        = "fal"
	ProviderDriverVolcengine = "volcengine"
)

// DbProvider stores configurable LLM provider metadata and credentials.
type DbProvider struct {
	ID          string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name        string    `gorm:"type:varchar(128);not null" json:"name"`
	Driver      string    `gorm:"type:varchar(64);not null" json:"driver"`
	Description string    `gorm:"type:text" json:"description"`
	APIKey      string    `gorm:"type:text" json:"api_key"`
	BaseURL     string    `gorm:"type:text" json:"base_url"`
	Config      JSONMap   `gorm:"type:json" json:"config"`
	IsActive    bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Models []DbModel `gorm:"foreignKey:ProviderID" json:"models,omitempty"`
}

// TableName overrides the table name for DbProvider.
func (DbProvider) TableName() string {
	return "llm_providers"
}

// DbModel stores provider-specific model configuration.
type DbModel struct {
	ID uint `gorm:"primarykey" json:"id"`

	ProviderID string `gorm:"column:provider_id;type:varchar(64);index:idx_provider_model,priority:1;not null" json:"provider_id"`
	ModelID    string `gorm:"column:model_id;type:varchar(255);index:idx_provider_model,priority:2;not null" json:"model_id"`

	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Price       string `gorm:"type:varchar(64)" json:"price"`

	MaxImages      int         `gorm:"column:max_images" json:"max_images"`
	Modalities     StringArray `gorm:"column:modalities;type:json" json:"modalities"`
	SupportedSizes StringArray `gorm:"column:supported_sizes;type:json" json:"supported_sizes"`
	Settings       JSONMap     `gorm:"column:settings;type:json" json:"settings"`

	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides the table name for DbModel.
func (DbModel) TableName() string {
	return "llm_models"
}

// ToLlmModel converts DbModel into LlmModel DTO.
func (m DbModel) ToLlmModel() LlmModel {
	var modalities []Modality
	for _, item := range m.Modalities {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			modalities = append(modalities, Modality(trimmed))
		}
	}

	inputs := Inputs{
		MaxImages: m.MaxImages,
	}
	if len(modalities) > 0 {
		inputs.Modalities = modalities
	}
	if len(m.SupportedSizes) > 0 {
		inputs.SupportedSizes = m.SupportedSizes.ToSlice()
	}

	return LlmModel{
		ID:          m.ModelID,
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Inputs:      inputs,
	}
}

// ToLlmProvider converts DbProvider with models into LlmProvider DTO.
func (p DbProvider) ToLlmProvider(models []DbModel) LlmProvider {
	out := LlmProvider{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
	}

	activeModels := make([]LlmModel, 0, len(models))
	for _, model := range models {
		if !model.IsActive {
			continue
		}
		activeModels = append(activeModels, model.ToLlmModel())
	}
	out.Models = activeModels
	return out
}
