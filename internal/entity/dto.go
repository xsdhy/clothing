package entity

type GenerateImageRequest struct {
	Prompt   string   `json:"prompt" binding:"required"`
	Images   []string `json:"images"`
	Size     string   `json:"size,omitempty"`
	Provider string   `json:"provider" binding:"required"`
	Model    string   `json:"model" binding:"required"`
}

type GenerateImageResponse struct {
	Image  string   `json:"image,omitempty"`
	Images []string `json:"images,omitempty"`
	Text   string   `json:"text,omitempty"`
}

type LlmModel struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       string   `json:"price"`
	Description string   `json:"description,omitempty"`
	ImageSizes  []string `json:"image_sizes,omitempty"`
}

type LlmProvider struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Models      []LlmModel `json:"models"`
}
