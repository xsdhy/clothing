package api

import (
	"clothing/internal/auth"
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/llm"
	"clothing/internal/model"
	"clothing/internal/storage"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type HTTPHandler struct {
	cfg               config.Config
	providers         []entity.LlmProvider
	providerMap       map[string]llm.AIService
	repo              model.Repository
	storage           storage.Storage
	storagePublicBase string
	authManager       *auth.Manager
}

func NewHTTPHandler(cfg config.Config, repo model.Repository, store storage.Storage) (*HTTPHandler, error) {
	providers, providerMap := NewProviderBundle(cfg)

	expiry := time.Duration(cfg.JWTExpirationMinutes) * time.Minute
	authManager, err := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, expiry)
	if err != nil {
		return nil, err
	}

	handler := &HTTPHandler{
		cfg:               cfg,
		providers:         providers,
		providerMap:       providerMap,
		repo:              repo,
		storage:           store,
		storagePublicBase: normalisePublicBase(cfg.StoragePublicBaseURL),
		authManager:       authManager,
	}

	return handler, nil
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

	dashscope, err := llm.NewDashscope(cfg)
	if err == nil && dashscope != nil {
		providerMap[dashscope.ProviderID()] = dashscope
		providers = append(providers, dashscope.Provider())
	} else if err != nil && strings.TrimSpace(cfg.DashscopeAPIKey) != "" {
		logrus.WithError(err).Warn("failed to initialise Dashscope provider")
	}

	falai, err := llm.NewFalAI(cfg)
	if err == nil && falai != nil {
		providerMap[falai.ProviderID()] = falai
		providers = append(providers, falai.Provider())
	} else if err != nil && strings.TrimSpace(cfg.FalAPIKey) != "" {
		logrus.WithError(err).Warn("failed to initialise fal.ai provider")
	}

	volcengine, err := llm.NewVolcengine(cfg)
	if err == nil && volcengine != nil {
		providerMap[volcengine.ProviderID()] = volcengine
		providers = append(providers, volcengine.Provider())
	} else if err != nil && strings.TrimSpace(cfg.VolcengineAPIKey) != "" {
		logrus.WithError(err).Warn("failed to initialise Volcengine provider")
	}

	return providers, providerMap
}

func normalisePublicBase(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = "/files"
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return strings.TrimRight(trimmed, "/")
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}
