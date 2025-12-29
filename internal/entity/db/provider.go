package db

import (
	"clothing/internal/entity/common"
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

// Provider 存储可配置的 LLM 服务商元数据和凭证。
type Provider struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	Driver      string         `gorm:"type:varchar(64);not null" json:"driver"`
	Description string         `gorm:"type:text" json:"description"`
	APIKey      string         `gorm:"type:text" json:"api_key"`
	BaseURL     string         `gorm:"type:text" json:"base_url"`
	Config      common.JSONMap `gorm:"type:json" json:"config"`
	IsActive    bool           `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	Models []Model `gorm:"foreignKey:ProviderID" json:"models,omitempty"`
}

// TableName 指定 Provider 的表名。
func (Provider) TableName() string {
	return "llm_providers"
}

// Model 存储服务商特定的模型配置。
type Model struct {
	ID uint `gorm:"primarykey" json:"id"`

	ProviderID string `gorm:"column:provider_id;type:varchar(64);index:idx_provider_model,priority:1;not null" json:"provider_id"`
	ModelID    string `gorm:"column:model_id;type:varchar(255);index:idx_provider_model,priority:2;not null" json:"model_id"`

	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Price       string `gorm:"type:varchar(64)" json:"price"`

	MaxImages int `gorm:"column:max_images" json:"max_images"`
	// Modalities 保持兼容旧字段，迁移后请使用 InputModalities/OutputModalities。
	Modalities         common.StringArray `gorm:"column:modalities;type:json" json:"-"`
	InputModalities    common.StringArray `gorm:"column:input_modalities;type:json" json:"input_modalities"`
	OutputModalities   common.StringArray `gorm:"column:output_modalities;type:json" json:"output_modalities"`
	SupportedSizes     common.StringArray `gorm:"column:supported_sizes;type:json" json:"supported_sizes"`
	SupportedDurations common.IntArray    `gorm:"column:supported_durations;type:json" json:"supported_durations"`
	DefaultSize        string             `gorm:"column:default_size;type:varchar(64)" json:"default_size"`
	DefaultDuration    int                `gorm:"column:default_duration" json:"default_duration"`
	Settings           common.JSONMap     `gorm:"column:settings;type:json" json:"settings"`

	// GenerationMode 指定内容生成方式（text_to_image, image_to_image, text_to_video, image_to_video）。
	// 替代 Settings["mode"]。
	GenerationMode string `gorm:"column:generation_mode;type:varchar(64)" json:"generation_mode"`
	// EndpointPath 是此模型的 API 端点路径。替代 Settings["endpoint"]。
	EndpointPath string `gorm:"column:endpoint_path;type:varchar(255)" json:"endpoint_path"`
	// SupportsStreaming 表示模型是否支持流式响应。
	SupportsStreaming bool `gorm:"column:supports_streaming;default:false" json:"supports_streaming"`
	// SupportsCancel 表示生成任务是否可以取消。
	SupportsCancel bool `gorm:"column:supports_cancel;default:false" json:"supports_cancel"`

	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定 Model 的表名。
func (Model) TableName() string {
	return "llm_models"
}

// IsVideoModel 检查此模型是否输出视频。
func (m *Model) IsVideoModel() bool {
	return m.OutputModalities.Contains("video")
}

// 解析设置的辅助函数

// ParseDurations 从设置中提取持续时间。
func ParseDurations(settings common.JSONMap) []int {
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

// MergeDurations 合并值与回退值，去除重复项。
func MergeDurations(values common.IntArray, fallback []int) []int {
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

// ResolveDefaultSize 从候选值、支持的尺寸和设置中解析默认尺寸。
func ResolveDefaultSize(candidate string, supported []string, settings common.JSONMap) string {
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

// ResolveDefaultDuration 从候选值、支持的持续时间和设置中解析默认持续时间。
func ResolveDefaultDuration(candidate int, supported []int, settings common.JSONMap) int {
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

func parseDefaultDuration(settings common.JSONMap) int {
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

func parseDefaultString(settings common.JSONMap, keys ...string) string {
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
