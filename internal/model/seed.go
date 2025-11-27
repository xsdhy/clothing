package model

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type providerSeed struct {
	Provider entity.DbProvider
	Models   []entity.DbModel
}

// SeedDefaultProviders ensures previously hard-coded providers/models exist in the database.
func SeedDefaultProviders(ctx context.Context, repo Repository, cfg config.Config) error {
	if repo == nil {
		return nil
	}

	seeds := buildDefaultProviderSeeds(cfg)
	for _, seed := range seeds {
		existing, err := repo.GetProvider(ctx, seed.Provider.ID)
		switch {
		case err == nil:
			if err := syncExistingProvider(ctx, repo, existing, seed); err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := createSeedProvider(ctx, repo, seed); err != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

func createSeedProvider(ctx context.Context, repo Repository, seed providerSeed) error {
	provider := seed.Provider
	provider.Models = nil

	if err := repo.CreateProvider(ctx, &provider); err != nil {
		return err
	}

	for _, modelSeed := range seed.Models {
		model := modelSeed
		model.ProviderID = provider.ID
		if err := repo.CreateModel(ctx, &model); err != nil {
			return err
		}
	}
	return nil
}

func syncExistingProvider(ctx context.Context, repo Repository, existing *entity.DbProvider, seed providerSeed) error {
	if existing == nil {
		return nil
	}

	updates := make(map[string]interface{})
	envAPIKey := strings.TrimSpace(seed.Provider.APIKey)
	if envAPIKey != "" && strings.TrimSpace(existing.APIKey) == "" {
		updates["api_key"] = envAPIKey
		if !existing.IsActive {
			updates["is_active"] = true
		}
	}
	if len(updates) > 0 {
		if err := repo.UpdateProvider(ctx, existing.ID, updates); err != nil {
			return err
		}
	}

	existingModelSet := make(map[string]struct{}, len(existing.Models))
	for _, model := range existing.Models {
		existingModelSet[strings.ToLower(strings.TrimSpace(model.ModelID))] = struct{}{}
	}

	for _, modelSeed := range seed.Models {
		key := strings.ToLower(strings.TrimSpace(modelSeed.ModelID))
		if _, ok := existingModelSet[key]; ok {
			continue
		}
		model := modelSeed
		model.ProviderID = existing.ID
		if err := repo.CreateModel(ctx, &model); err != nil {
			return err
		}
	}
	return nil
}

func buildDefaultProviderSeeds(cfg config.Config) []providerSeed {
	openRouterKey := strings.TrimSpace(cfg.OpenRouterAPIKey)
	geminiKey := strings.TrimSpace(cfg.GeminiAPIKey)
	aiHubMixKey := strings.TrimSpace(cfg.AiHubMixAPIKey)
	dashscopeKey := strings.TrimSpace(cfg.DashscopeAPIKey)
	falKey := strings.TrimSpace(cfg.FalAPIKey)
	volcengineKey := strings.TrimSpace(cfg.VolcengineAPIKey)

	return []providerSeed{
		{
			Provider: entity.DbProvider{
				ID:       "openrouter",
				Name:     "OpenRouter",
				Driver:   entity.ProviderDriverOpenRouter,
				BaseURL:  "https://openrouter.ai/api/v1/chat/completions",
				APIKey:   openRouterKey,
				IsActive: openRouterKey != "",
			},
			Models: []entity.DbModel{
				{
					ModelID:  "google/gemini-2.5-flash-image-preview",
					Name:     "Gemini 2.5 Flash Image Preview",
					IsActive: true,
				},
			},
		},
		{
			Provider: entity.DbProvider{
				ID:       "google",
				Name:     "Google Gemini",
				Driver:   entity.ProviderDriverGemini,
				APIKey:   geminiKey,
				IsActive: geminiKey != "",
			},
			Models: []entity.DbModel{
				{
					ModelID:  "gemini-2.5-flash-image-preview",
					Name:     "Gemini 2.5 Flash Image Preview",
					IsActive: true,
				},
			},
		},
		{
			Provider: entity.DbProvider{
				ID:       "aihubmix",
				Name:     "AiHubMix",
				Driver:   entity.ProviderDriverAiHubMix,
				BaseURL:  "https://aihubmix.com/v1/chat/completions",
				APIKey:   aiHubMixKey,
				IsActive: aiHubMixKey != "",
			},
			Models: []entity.DbModel{
				{
					ModelID:  "qwen-image-edit",
					Name:     "Qwen Image Edit",
					Price:    "<UNK>",
					IsActive: true,
				},
				{
					ModelID:          "qwen3-vl-235b-a22b-instruct",
					Name:             "Qwen3-VL-235B-A22B Instruct",
					Description:      "视频/图像理解多模态模型",
					Price:            "<UNK>",
					InputModalities:  entity.StringArray{"text", "image", "video"},
					OutputModalities: entity.StringArray{"text"},
					IsActive:         true,
				},
				{
					ModelID:  "gemini-2.5-flash-image-preview",
					Name:     "Gemini 2.5 Flash Image Preview",
					Price:    "0.03/IMG",
					IsActive: true,
				},
				{
					ModelID:  "imagen-4.0-fast-generate-001",
					Name:     "Imagen 4.0 Fast Generate",
					Price:    "0.03/IMG",
					IsActive: true,
				},
				{
					ModelID:  "gpt-4o-image",
					Name:     "GPT-4o Image",
					Price:    "0.005/IMG",
					IsActive: true,
				},
			},
		},
		{
			Provider: entity.DbProvider{
				ID:       "dashscope",
				Name:     "Dashscope",
				Driver:   entity.ProviderDriverDashscope,
				APIKey:   dashscopeKey,
				IsActive: dashscopeKey != "",
			},
			Models: []entity.DbModel{
				{
					ModelID:  "qwen-image-edit",
					Name:     "Qwen Image Edit",
					IsActive: true,
				},
			},
		},
		{
			Provider: entity.DbProvider{
				ID:       "fal",
				Name:     "fal.ai",
				Driver:   entity.ProviderDriverFal,
				BaseURL:  "https://fal.run",
				APIKey:   falKey,
				IsActive: falKey != "",
			},
			Models: []entity.DbModel{
				{
					ModelID:          "fal-ai/hunyuan-image/v3/text-to-image",
					Name:             "Hunyuan Image v3",
					Price:            "$0.1",
					IsActive:         true,
					InputModalities:  entity.StringArray{"text"},
					OutputModalities: entity.StringArray{"image"},
					Settings: entity.JSONMap{
						"endpoint": "/fal-ai/hunyuan-image/v3/text-to-image",
						"mode":     "text_to_image",
					},
				},
				{
					ModelID:          "fal-ai/qwen-image-edit/image-to-image",
					Name:             "Qwen Image Edit",
					Price:            "$0.03",
					IsActive:         true,
					InputModalities:  entity.StringArray{"text", "image"},
					OutputModalities: entity.StringArray{"image"},
					Settings: entity.JSONMap{
						"endpoint": "/fal-ai/qwen-image-edit/image-to-image",
						"mode":     "image_to_image",
					},
				},
			},
		},
		{
			Provider: entity.DbProvider{
				ID:       "volcengine",
				Name:     "Volcengine",
				Driver:   entity.ProviderDriverVolcengine,
				APIKey:   volcengineKey,
				IsActive: volcengineKey != "",
			},
			Models: []entity.DbModel{
				{
					ModelID:          "doubao-seedream-4-0-250828",
					Name:             "Doubao Seedream 4.0",
					Description:      "火山引擎图像生成模型",
					IsActive:         true,
					MaxImages:        9,
					InputModalities:  entity.StringArray{"text", "image"},
					OutputModalities: entity.StringArray{"image"},
					SupportedSizes:   entity.StringArray{"1K", "2K", "4K"},
				},
			},
		},
	}
}
