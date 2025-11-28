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
	ClientID string `json:"client_id,omitempty"` // 客户端ID，结束后SSE推送使用

	Prompt     string         `json:"prompt" binding:"required"`
	Inputs     ContentInputs  `json:"inputs"`
	Options    ContentOptions `json:"options"`
	ProviderID string         `json:"provider_id" binding:"required"` // 供应商ID
	ModelID    string         `json:"model_id" binding:"required"`    // 模型ID
	TagIDs     []uint         `json:"tag_ids"`
}

type GenerateContentResponse struct {
	ImageAssets      []string `json:"image_assets,omitempty"`
	TextContent      string   `json:"text_content,omitempty"`
	ExternalTaskCode string   `json:"external_task_code,omitempty"`
	RequestID        string   `json:"request_id,omitempty"`
}

// 模态枚举（可扩展到音频/视频）
type Modality string

const (
	ModText  Modality = "text"
	ModImage Modality = "image"
	ModVideo Modality = "video"
)

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
