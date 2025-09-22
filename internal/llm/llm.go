package llm

import (
	"clothing/internal/entity"
	"context"
)

type AIService interface {
	ProviderID() string
	Provider() entity.LlmProvider
	Models() []entity.LlmModel
	SupportsModel(modelID string) bool

	GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error)
}
