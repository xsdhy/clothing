package dto

import "strings"

// MediaInput represents a unified media input with type and role information.
type MediaInput struct {
	Type    string `json:"type"`              // image, video
	Content string `json:"content"`           // URL, Base64, or DataURL
	Role    string `json:"role,omitempty"`    // reference, first_frame, last_frame
}

// OutputConfig contains output configuration for generation.
type OutputConfig struct {
	Size       string `json:"size,omitempty"`
	Duration   int    `json:"duration,omitempty"`
	NumOutputs int    `json:"num_outputs,omitempty"`
}

// MediaOutput represents a unified media output.
type MediaOutput struct {
	Type     string `json:"type"`               // image, video
	URL      string `json:"url"`
	MimeType string `json:"mime_type,omitempty"`
}

// GenerateContentRequest is the request payload for content generation.
type GenerateContentRequest struct {
	ClientID   string `json:"client_id,omitempty"` // 客户端ID，结束后SSE推送使用
	ProviderID string `json:"provider_id" binding:"required"` // 供应商ID
	ModelID    string `json:"model_id" binding:"required"`    // 模型ID

	Prompt string `json:"prompt" binding:"required"`

	// New unified input format
	InputMedia []MediaInput `json:"input_media,omitempty"`

	// New output configuration
	Output OutputConfig `json:"output,omitempty"`

	TagIDs []uint `json:"tag_ids,omitempty"`
}

// GetImages returns all image inputs from InputMedia.
func (r *GenerateContentRequest) GetImages() []string {
	images := make([]string, 0, len(r.InputMedia))
	for _, m := range r.InputMedia {
		if !strings.EqualFold(strings.TrimSpace(m.Type), "image") {
			continue
		}
		content := strings.TrimSpace(m.Content)
		if content != "" {
			images = append(images, content)
		}
	}
	return images
}

// GetVideos returns all video inputs from InputMedia.
func (r *GenerateContentRequest) GetVideos() []string {
	videos := make([]string, 0, len(r.InputMedia))
	for _, m := range r.InputMedia {
		if !strings.EqualFold(strings.TrimSpace(m.Type), "video") {
			continue
		}
		content := strings.TrimSpace(m.Content)
		if content != "" {
			videos = append(videos, content)
		}
	}
	return videos
}

// GetSize returns the output size.
func (r *GenerateContentRequest) GetSize() string {
	return r.Output.Size
}

// GetDuration returns the output duration.
func (r *GenerateContentRequest) GetDuration() int {
	return r.Output.Duration
}

// GenerateContentResponse is the response from content generation.
type GenerateContentResponse struct {
	// New unified output format
	Outputs []MediaOutput `json:"outputs,omitempty"`
	Text    string        `json:"text,omitempty"`

	// Task identification
	TaskID    string `json:"task_id,omitempty"`     // Unified task ID
	RequestID string `json:"request_id,omitempty"`
}
