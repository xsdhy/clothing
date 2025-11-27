package entity

import "time"

type ContentInputs struct {
	Images []string `json:"images,omitempty"`
	Videos []string `json:"videos,omitempty"`
}

type ContentOptions struct {
	Size     string `json:"size,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

type GenerateContentRequest struct {
	Prompt     string         `json:"prompt" binding:"required"`
	Inputs     ContentInputs  `json:"inputs"`
	Options    ContentOptions `json:"options"`
	ProviderID string         `json:"provider_id" binding:"required"` // 供应商ID
	ModelID    string         `json:"model_id" binding:"required"`    // 模型ID
	TagIDs     []uint         `json:"tag_ids"`
}

type GenerateContentResponse struct {
	Outputs []string `json:"outputs,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// 模态枚举（可扩展到音频/视频）
type Modality string

const (
	ModText  Modality = "text"
	ModImage Modality = "image"
	ModVideo Modality = "video"
)

type Inputs struct {
	InputModalities    []Modality `json:"input_modalities,omitempty"`    // 支持的输入模态：text/image/video
	OutputModalities   []Modality `json:"output_modalities,omitempty"`   // 支持的输出模态：text/image/video
	MaxImages          int        `json:"max_images,omitempty"`          // 支持的最大输入图片数
	SupportedSizes     []string   `json:"supported_sizes,omitempty"`     // 支持的图像尺寸/分辨率，留空表示不限预设
	SupportedDurations []int      `json:"supported_durations,omitempty"` // 视频模型支持的时长（秒）
	DefaultSize        string     `json:"default_size,omitempty"`        // 尺寸/分辨率默认值
	DefaultDuration    int        `json:"default_duration,omitempty"`    // 视频默认时长（秒）
}

type LlmModel struct {
	ID          string `json:"id"`                    // 模型ID
	Name        string `json:"name"`                  //显示名称
	Price       string `json:"price"`                 // 价格
	Description string `json:"description,omitempty"` //描述

	Inputs Inputs `json:"inputs,omitempty"` //输入的配置、模型的能力等
}

type LlmProvider struct {
	ID          string     `json:"id"`                    //供应商ID
	Name        string     `json:"name"`                  //显示名称
	Description string     `json:"description,omitempty"` //描述
	Models      []LlmModel `json:"models"`
}

type UsageRecordQuery struct {
	BaseParams
	Provider        string `json:"provider" form:"provider" query:"provider"`
	Model           string `json:"model" form:"model" query:"model"`
	Result          string `json:"result" form:"result" query:"result"`
	UserID          uint   `json:"-" form:"-" query:"-"`
	IncludeAll      bool   `json:"-" form:"-" query:"-"`
	TagIDs          []uint `json:"-" form:"-" query:"-"`
	HasOutputImages bool   `json:"-" form:"has_output_images" query:"has_output_images"`
}

type UsageImage struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

type UsageRecordItem struct {
	ID           uint         `json:"id"`
	ProviderID   string       `json:"provider_id"`
	ModelID      string       `json:"model_id"`
	Prompt       string       `json:"prompt"`
	Size         string       `json:"size"`
	OutputText   string       `json:"output_text"`
	ErrorMessage string       `json:"error_message"`
	CreatedAt    time.Time    `json:"created_at"`
	InputImages  []UsageImage `json:"input_images"`
	OutputImages []UsageImage `json:"output_images"`
	User         UserSummary  `json:"user"`
	Tags         []Tag        `json:"tags"`
}

type UsageRecordListResponse struct {
	Records []UsageRecordItem `json:"records"`
	Meta    *Meta             `json:"meta"`
}

type UsageRecordDetailResponse struct {
	Record UsageRecordItem `json:"record"`
}

type Tag struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	UsageCount  int64     `json:"usage_count,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Description string    `json:"description,omitempty"`
}

type TagListResponse struct {
	Tags []Tag `json:"tags"`
}

type TagDetailResponse struct {
	Tag Tag `json:"tag"`
}
