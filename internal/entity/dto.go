package entity

import "time"

type GenerateImageRequest struct {
	Prompt   string   `json:"prompt" binding:"required"`
	Images   []string `json:"images"`
	Size     string   `json:"size,omitempty"`
	Provider string   `json:"provider" binding:"required"` //供应商ID
	Model    string   `json:"model" binding:"required"`    //模型ID
}

type GenerateImageResponse struct {
	Images []string `json:"images,omitempty"`
	Text   string   `json:"text,omitempty"`
}

// 模态枚举（可扩展到音频/视频）
type Modality string

const (
	ModText  Modality = "text"
	ModImage Modality = "image"
)

type Inputs struct {
	Modalities     []Modality `json:"modalities,omitempty"`      // 支持的模态枚举，例: ["text"], ["image"], ["text","image"]
	MaxImages      int        `json:"max_images,omitempty"`      // 支持的最大输入图片数
	SupportedSizes []string   `json:"supported_sizes,omitempty"` // 支持的图像尺寸，留空表示不限预设
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
	Provider string `json:"provider" form:"provider" query:"provider"`
	Model    string `json:"model" form:"model" query:"model"`
	Result   string `json:"result" form:"result" query:"result"`
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
}

type UsageRecordListResponse struct {
	Records []UsageRecordItem `json:"records"`
	Meta    *Meta             `json:"meta"`
}
