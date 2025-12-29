package dto

import (
	"clothing/internal/entity/common"
	"time"
)

// UsageRecordQuery supports querying usage records.
type UsageRecordQuery struct {
	common.BaseParams
	Provider        string `json:"provider" form:"provider" query:"provider"`
	Model           string `json:"model" form:"model" query:"model"`
	Result          string `json:"result" form:"result" query:"result"`
	UserID          uint   `json:"-" form:"-" query:"-"`
	IncludeAll      bool   `json:"-" form:"-" query:"-"`
	TagIDs          []uint `json:"-" form:"-" query:"-"`
	HasOutputImages bool   `json:"-" form:"has_output_images" query:"has_output_images"`
}

// UsageImage represents an image in usage records.
type UsageImage struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

// UsageRecordItem is the response representation of a usage record.
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

// UsageRecordListResponse is the response for listing usage records.
type UsageRecordListResponse struct {
	Records []UsageRecordItem `json:"records"`
	Meta    *common.Meta      `json:"meta"`
}

// UsageRecordDetailResponse is the response for a single usage record.
type UsageRecordDetailResponse struct {
	Record UsageRecordItem `json:"record"`
}
