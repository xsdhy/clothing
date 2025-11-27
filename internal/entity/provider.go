package entity

import (
	"strconv"
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

	MaxImages          int         `gorm:"column:max_images" json:"max_images"`
	Modalities         StringArray `gorm:"column:modalities;type:json" json:"modalities"`
	SupportedSizes     StringArray `gorm:"column:supported_sizes;type:json" json:"supported_sizes"`
	SupportedDurations IntArray    `gorm:"column:supported_durations;type:json" json:"supported_durations"`
	DefaultSize        string      `gorm:"column:default_size;type:varchar(64)" json:"default_size"`
	DefaultDuration    int         `gorm:"column:default_duration" json:"default_duration"`
	Settings           JSONMap     `gorm:"column:settings;type:json" json:"settings"`

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

	durations := mergeDurations(m.SupportedDurations, parseDurations(m.Settings))
	if len(durations) > 0 {
		inputs.SupportedDurations = durations
	}

	if defaultSize := resolveDefaultSize(m.DefaultSize, inputs.SupportedSizes, m.Settings); defaultSize != "" {
		inputs.DefaultSize = defaultSize
	}

	if defaultDuration := resolveDefaultDuration(m.DefaultDuration, inputs.SupportedDurations, m.Settings); defaultDuration > 0 {
		inputs.DefaultDuration = defaultDuration
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

func parseDurations(settings JSONMap) []int {
	raw := settings["durations"]
	var out []int
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			switch vv := item.(type) {
			case float64:
				out = append(out, int(vv))
			case int:
				out = append(out, vv)
			case int64:
				out = append(out, int(vv))
			case string:
				trim := strings.TrimSpace(vv)
				if trim == "" {
					continue
				}
				if num, err := strconv.Atoi(trim); err == nil {
					out = append(out, num)
				}
			}
		}
	case []int:
		out = append(out, v...)
	case []float64:
		for _, item := range v {
			out = append(out, int(item))
		}
	}
	return out
}

func mergeDurations(values IntArray, fallback []int) []int {
	out := make([]int, 0, len(values)+len(fallback))
	seen := make(map[int]struct{})

	add := func(v int) {
		if v <= 0 {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	for _, v := range values {
		add(v)
	}
	for _, v := range fallback {
		add(v)
	}

	return out
}

func resolveDefaultSize(candidate string, supported []string, settings JSONMap) string {
	defaultSize := strings.TrimSpace(candidate)
	if defaultSize == "" {
		defaultSize = parseDefaultString(settings, "default_size", "default_resolution")
	}
	if len(supported) == 0 {
		return defaultSize
	}
	for _, size := range supported {
		if strings.EqualFold(defaultSize, size) {
			return size
		}
	}
	return supported[0]
}

func resolveDefaultDuration(candidate int, supported []int, settings JSONMap) int {
	defaultDuration := candidate
	if defaultDuration <= 0 {
		defaultDuration = parseDefaultDuration(settings)
	}
	if len(supported) == 0 {
		return defaultDuration
	}
	for _, duration := range supported {
		if duration == defaultDuration {
			return duration
		}
	}
	return supported[0]
}

func parseDefaultDuration(settings JSONMap) int {
	raw := settings["default_duration"]
	switch v := raw.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		trim := strings.TrimSpace(v)
		if trim == "" {
			return 0
		}
		if num, err := strconv.Atoi(trim); err == nil {
			return num
		}
	}
	return 0
}

func parseDefaultString(settings JSONMap, keys ...string) string {
	for _, key := range keys {
		if raw, ok := settings[key]; ok {
			switch v := raw.(type) {
			case string:
				if trimmed := strings.TrimSpace(v); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return ""
}
