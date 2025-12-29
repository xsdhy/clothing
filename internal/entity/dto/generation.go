package dto

// ContentInputs contains input media for generation requests.
// Deprecated: Use InputMedia instead for new code.
type ContentInputs struct {
	Images []string `json:"images,omitempty"`
	Videos []string `json:"videos,omitempty"`
}

// ContentOptions contains generation options.
// Deprecated: Use OutputConfig instead for new code.
type ContentOptions struct {
	Size     string `json:"size,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

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

	// Legacy fields for backward compatibility
	Inputs  ContentInputs  `json:"inputs,omitempty"`
	Options ContentOptions `json:"options,omitempty"`

	TagIDs []uint `json:"tag_ids,omitempty"`
}

// NormalizeInputs converts legacy input formats to the new InputMedia format.
// Call this before processing the request to ensure consistent handling.
func (r *GenerateContentRequest) NormalizeInputs() {
	// Convert legacy Inputs.Images to InputMedia
	if len(r.InputMedia) == 0 && len(r.Inputs.Images) > 0 {
		for _, img := range r.Inputs.Images {
			r.InputMedia = append(r.InputMedia, MediaInput{
				Type:    "image",
				Content: img,
			})
		}
	}

	// Convert legacy Inputs.Videos to InputMedia
	if len(r.Inputs.Videos) > 0 {
		for _, vid := range r.Inputs.Videos {
			r.InputMedia = append(r.InputMedia, MediaInput{
				Type:    "video",
				Content: vid,
			})
		}
	}

	// Merge legacy Options into Output
	if r.Output.Size == "" && r.Options.Size != "" {
		r.Output.Size = r.Options.Size
	}
	if r.Output.Duration == 0 && r.Options.Duration > 0 {
		r.Output.Duration = r.Options.Duration
	}
}

// GetImages returns all image inputs from InputMedia for compatibility.
func (r *GenerateContentRequest) GetImages() []string {
	var images []string
	for _, m := range r.InputMedia {
		if m.Type == "image" || m.Type == "" {
			images = append(images, m.Content)
		}
	}
	// Fallback to legacy format
	if len(images) == 0 {
		images = r.Inputs.Images
	}
	return images
}

// GetVideos returns all video inputs from InputMedia for compatibility.
func (r *GenerateContentRequest) GetVideos() []string {
	var videos []string
	for _, m := range r.InputMedia {
		if m.Type == "video" {
			videos = append(videos, m.Content)
		}
	}
	// Fallback to legacy format
	if len(videos) == 0 {
		videos = r.Inputs.Videos
	}
	return videos
}

// GetSize returns the output size from Output or Options.
func (r *GenerateContentRequest) GetSize() string {
	if r.Output.Size != "" {
		return r.Output.Size
	}
	return r.Options.Size
}

// GetDuration returns the output duration from Output or Options.
func (r *GenerateContentRequest) GetDuration() int {
	if r.Output.Duration > 0 {
		return r.Output.Duration
	}
	return r.Options.Duration
}

// GenerateContentResponse is the response from content generation.
type GenerateContentResponse struct {
	// New unified output format
	Outputs []MediaOutput `json:"outputs,omitempty"`
	Text    string        `json:"text,omitempty"`

	// Legacy fields for backward compatibility
	ImageAssets []string `json:"image_assets,omitempty"`
	TextContent string   `json:"text_content,omitempty"`

	// Task identification
	TaskID    string `json:"task_id,omitempty"`     // Unified task ID
	RequestID string `json:"request_id,omitempty"`

	// Deprecated: Use TaskID instead
	ExternalTaskCode string `json:"external_task_code,omitempty"`
}

// ToLegacyFormat populates legacy fields from new format for backward compatibility.
// Call this before returning the response to ensure old clients work correctly.
func (r *GenerateContentResponse) ToLegacyFormat() {
	// Copy Outputs to ImageAssets
	if len(r.Outputs) > 0 && len(r.ImageAssets) == 0 {
		for _, o := range r.Outputs {
			r.ImageAssets = append(r.ImageAssets, o.URL)
		}
	}

	// Copy Text to TextContent
	if r.Text != "" && r.TextContent == "" {
		r.TextContent = r.Text
	}

	// Copy TaskID to ExternalTaskCode
	if r.TaskID != "" && r.ExternalTaskCode == "" {
		r.ExternalTaskCode = r.TaskID
	}
}

// NormalizeFromLegacy populates new format from legacy fields.
func (r *GenerateContentResponse) NormalizeFromLegacy() {
	// Copy ImageAssets to Outputs
	if len(r.ImageAssets) > 0 && len(r.Outputs) == 0 {
		for _, url := range r.ImageAssets {
			mediaType := "image"
			// Simple heuristic for video detection
			if isVideoURL(url) {
				mediaType = "video"
			}
			r.Outputs = append(r.Outputs, MediaOutput{
				Type: mediaType,
				URL:  url,
			})
		}
	}

	// Copy TextContent to Text
	if r.TextContent != "" && r.Text == "" {
		r.Text = r.TextContent
	}

	// Copy ExternalTaskCode to TaskID
	if r.ExternalTaskCode != "" && r.TaskID == "" {
		r.TaskID = r.ExternalTaskCode
	}
}

// isVideoURL performs a simple check to detect video URLs.
func isVideoURL(url string) bool {
	lower := toLower(url)
	return contains(lower, ".mp4") || contains(lower, ".mov") || contains(lower, ".webm") ||
		contains(lower, ".avi") || contains(lower, "video")
}

// toLower is a simple lowercase helper.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
