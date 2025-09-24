package llm

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"context"
	"errors"
	"net/http"
	"strings"
)

type GeminiService struct {
	httpClient   *http.Client
	GeminiAPIKey string

	modelLookup map[string]struct{}
	models      []entity.LlmModel
}

func NewGeminiService(cfg config.Config) (*GeminiService, error) {
	if strings.TrimSpace(cfg.GeminiAPIKey) == "" {
		return nil, errors.New("gemini api key is not configured")
	}

	httpClient := &http.Client{}

	models := []entity.LlmModel{
		{
			ID:          "gemini-2.5-flash-image-preview",
			Name:        "Gemini 2.5 Flash Image Preview",
			Description: "最新的实验性模型，支持图像生成",
		},
	}

	g := &GeminiService{
		httpClient:   httpClient,
		GeminiAPIKey: cfg.GeminiAPIKey,
		models:       models,
		modelLookup: func() map[string]struct{} {
			lookup := make(map[string]struct{}, len(models))
			for _, model := range models {
				lookup[model.ID] = struct{}{}
			}
			return lookup
		}(),
	}

	return g, nil
}

func (g *GeminiService) ProviderID() string {
	return "gemini"
}

func (g *GeminiService) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     "google",
		Name:   "Google Gemini",
		Models: g.Models(),
	}
}
func (g *GeminiService) Models() []entity.LlmModel {
	return g.models
}
func (g *GeminiService) SupportsModel(modelID string) bool {
	if g == nil || modelID == "" {
		return false
	}
	_, ok := g.modelLookup[modelID]
	return ok
}

func (g *GeminiService) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {

	//todo:: 原生请求
	return "", "", errors.New("gemini response did not include image data")
}
