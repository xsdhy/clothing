package api

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/llm"
	"strings"

	"github.com/sirupsen/logrus"
)

type HTTPHandler struct {
	cfg         config.Config
	providers   []entity.LlmProvider
	providerMap map[string]llm.AIService
}

func NewHTTPHandler(cfg config.Config) *HTTPHandler {
	providers, providerMap := NewProviderBundle(cfg)

	return &HTTPHandler{
		cfg:         cfg,
		providers:   providers,
		providerMap: providerMap,
	}
}

func NewProviderBundle(cfg config.Config) ([]entity.LlmProvider, map[string]llm.AIService) {
	providers := make([]entity.LlmProvider, 0)
	providerMap := make(map[string]llm.AIService)

	openRouter, err := llm.NewOpenRouter(cfg)
	if err == nil && openRouter != nil {
		providerMap[openRouter.ProviderID()] = openRouter
		providers = append(providers, openRouter.Provider())
	} else if err != nil && strings.TrimSpace(cfg.OpenRouterAPIKey) != "" {
		logrus.WithError(err).Warn("failed to initialise OpenRouter provider")
	}

	gemini, err := llm.NewGeminiService(cfg)
	if err == nil && gemini != nil {
		providerMap[gemini.ProviderID()] = gemini
		providers = append(providers, gemini.Provider())
	} else if err != nil && strings.TrimSpace(cfg.GeminiAPIKey) != "" {
		logrus.WithError(err).Warn("failed to initialise Gemini provider")
	}

	aiHubMix, err := llm.NewAiHubMix(cfg)
	if err == nil && aiHubMix != nil {
		providerMap[aiHubMix.ProviderID()] = aiHubMix
		providers = append(providers, aiHubMix.Provider())
	} else if err != nil && strings.TrimSpace(cfg.AiHubMixAPIKey) != "" {
		logrus.WithError(err).Warn("failed to initialise Gemini provider")
	}

	return providers, providerMap
}
