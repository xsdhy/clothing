package api

import (
	"clothing/internal/auth"
	"clothing/internal/config"
	"clothing/internal/model"
	"clothing/internal/service"
	"clothing/internal/storage"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPHandler HTTP 请求处理器
type HTTPHandler struct {
	cfg               config.Config
	repo              model.Repository
	storage           storage.Storage
	storagePublicBase string
	authManager       *auth.Manager

	// 服务层
	generationService *service.GenerationService

	// SSE 客户端管理
	sseClients map[string][]chan sseMessage
	sseMu      sync.Mutex
}

// NewHTTPHandler 创建 HTTP 处理器实例
func NewHTTPHandler(cfg config.Config, repo model.Repository, store storage.Storage) (*HTTPHandler, error) {
	expiry := time.Duration(cfg.JWTExpirationMinutes) * time.Minute
	authManager, err := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, expiry)
	if err != nil {
		return nil, err
	}

	// 创建生成服务
	generationSvc := service.NewGenerationService(repo, store)

	handler := &HTTPHandler{
		cfg:               cfg,
		repo:              repo,
		storage:           store,
		storagePublicBase: normalisePublicBase(cfg.StoragePublicBaseURL),
		authManager:       authManager,
		generationService: generationSvc,
		sseClients:        make(map[string][]chan sseMessage),
	}

	// 设置 SSE 通知回调
	generationSvc.SetNotifyFunc(handler.notifyGenerationComplete)

	return handler, nil
}

// normalisePublicBase 规范化公共 URL 基础路径
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

// notifyGenerationComplete 通知生成完成（用于 SSE 推送）
func (h *HTTPHandler) notifyGenerationComplete(clientID string, recordID uint, status string, errMsg string) {
	if strings.TrimSpace(clientID) == "" {
		return
	}
	payload := gin.H{
		"record_id": recordID,
		"status":    status,
	}
	if trimmed := strings.TrimSpace(errMsg); trimmed != "" {
		payload["error"] = trimmed
	}
	h.publishSSEMessage(clientID, sseMessage{
		event: "generation_completed",
		data:  payload,
	})
}
