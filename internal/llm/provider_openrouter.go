package llm

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/utils"
	"context"
	"errors"
	"github.com/sashabaranov/go-openai"
	"strings"
)

type OpenRouter struct {
	client *openai.Client

	modelLookup map[string]struct{}
}

func NewOpenRouter(cfg config.Config) (*OpenRouter, error) {
	openAIConfig := openai.DefaultConfig(cfg.OpenRouterAPIKey)
	openAIConfig.BaseURL = "https://openrouter.ai/api/v1"

	o := &OpenRouter{
		client: openai.NewClientWithConfig(openAIConfig),
	}

	models := o.Models()
	for _, model := range models {
		o.modelLookup[model.ID] = struct{}{}
	}

	return o, nil
}
func (o *OpenRouter) ProviderID() string {
	return "openrouter"
}

func (o *OpenRouter) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     "openrouter",
		Name:   "OpenRouter",
		Models: o.Models(),
	}
}

func (o *OpenRouter) Models() []entity.LlmModel {
	return []entity.LlmModel{
		{
			ID:          "google/gemini-2.5-flash-image-preview:free",
			Name:        "Gemini 2.5 Flash Image (Free)",
			Description: "免费的图像生成模型",
		},
		{
			ID:          "google/gemini-2.5-flash-image-preview",
			Name:        "Gemini 2.5 Flash Image Preview",
			Description: "Google 的高性能模型",
		},
		{
			ID:          "openai/gpt-4o",
			Name:        "GPT-4o",
			Description: "OpenAI 的多模态模型",
		},
		{
			ID:          "meta-llama/llama-3.2-11b-vision-instruct",
			Name:        "Llama 3.2 Vision",
			Description: "Meta 的视觉理解模型",
		},
	}

}

func (o *OpenRouter) SupportsModel(modelID string) bool {
	if o == nil || modelID == "" {
		return false
	}
	_, ok := o.modelLookup[modelID]
	return ok
}

func (o *OpenRouter) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {
	parts := []openai.ChatMessagePart{
		{
			Type: openai.ChatMessagePartTypeText,
			Text: request.Prompt,
		},
	}

	for _, image := range request.Images {
		image = strings.TrimSpace(image)
		if image == "" {
			continue
		}
		parts = append(parts, openai.ChatMessagePart{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: utils.EnsureDataURL(image)},
		})
	}

	completion, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: request.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:         openai.ChatMessageRoleUser,
				MultiContent: parts,
			},
		},
		MaxTokens: 1000,
	})
	if err != nil {
		return "", "", err
	}

	for _, choice := range completion.Choices {
		if len(choice.Message.MultiContent) > 0 {
			for _, part := range choice.Message.MultiContent {
				if part.Type == openai.ChatMessagePartTypeImageURL && part.ImageURL != nil && part.ImageURL.URL != "" {
					return part.ImageURL.URL, "", nil
				}
				if part.Text != "" {
					return "", part.Text, nil
				}
			}
		}
		if choice.Message.Content != "" {
			return "", choice.Message.Content, nil
		}
	}

	return "", "", errors.New("model did not return an image")
}
