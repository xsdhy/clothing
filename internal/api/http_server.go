package api

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/llm"
	"net/http"
)

type headerTransport struct {
	headers http.Header
	base    http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, values := range t.headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return t.base.RoundTrip(req)
}

type HTTPHandler struct {
	cfg        config.Config
	httpClient *http.Client

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
	}

	gemini, err := llm.NewGeminiService(cfg)
	if err == nil && gemini != nil {
		providerMap[gemini.ProviderID()] = gemini
		providers = append(providers, gemini.Provider())
	}

	return providers, providerMap
}
