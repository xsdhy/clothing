package entity

type GenerateImageRequest struct {
	Prompt   string   `json:"prompt" binding:"required"`
	Images   []string `json:"images"`
	Provider string   `json:"provider" binding:"required"`
	Model    string   `json:"model" binding:"required"`
}

type GenerateImageResponse struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

type LlmModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description,omitempty"`
}

type LlmProvider struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Models      []LlmModel `json:"models"`
}
