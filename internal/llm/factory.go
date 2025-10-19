package llm

import (
	"clothing/internal/entity"
	"fmt"
	"strings"
)

// NewService instantiates an AIService implementation for a provider.
func NewService(provider *entity.DbProvider) (AIService, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider config is nil")
	}

	driver := strings.ToLower(strings.TrimSpace(provider.Driver))
	if driver == "" {
		driver = strings.ToLower(strings.TrimSpace(provider.ID))
	}

	models := provider.Models

	switch driver {
	case entity.ProviderDriverOpenRouter:
		return NewOpenRouter(provider, models)
	case entity.ProviderDriverGemini:
		return NewGeminiService(provider, models)
	case entity.ProviderDriverAiHubMix:
		return NewAiHubMix(provider, models)
	case entity.ProviderDriverDashscope:
		return NewDashscope(provider, models)
	case entity.ProviderDriverFal:
		return NewFalAI(provider, models)
	case entity.ProviderDriverVolcengine:
		return NewVolcengine(provider, models)
	default:
		return nil, fmt.Errorf("unsupported provider driver: %s", provider.Driver)
	}
}
