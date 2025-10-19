package api

import (
	"clothing/internal/auth"
	"clothing/internal/config"
	"clothing/internal/model"
	"clothing/internal/storage"
	"strings"
	"time"
)

type HTTPHandler struct {
	cfg               config.Config
	repo              model.Repository
	storage           storage.Storage
	storagePublicBase string
	authManager       *auth.Manager
}

func NewHTTPHandler(cfg config.Config, repo model.Repository, store storage.Storage) (*HTTPHandler, error) {
	expiry := time.Duration(cfg.JWTExpirationMinutes) * time.Minute
	authManager, err := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, expiry)
	if err != nil {
		return nil, err
	}

	handler := &HTTPHandler{
		cfg:               cfg,
		repo:              repo,
		storage:           store,
		storagePublicBase: normalisePublicBase(cfg.StoragePublicBaseURL),
		authManager:       authManager,
	}

	return handler, nil
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
