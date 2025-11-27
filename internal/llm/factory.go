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

	switch driver {
	case entity.ProviderDriverOpenRouter:
		return NewOpenRouter(provider)
	case entity.ProviderDriverGemini:
		return NewGeminiService(provider)
	case entity.ProviderDriverAiHubMix:
		return NewAiHubMix(provider)
	case entity.ProviderDriverDashscope:
		return NewDashscope(provider)
	case entity.ProviderDriverFal:
		return NewFalAI(provider)
	case entity.ProviderDriverVolcengine:
		return NewVolcengine(provider)
	default:
		return nil, fmt.Errorf("unsupported provider driver: %s", provider.Driver)
	}
}
