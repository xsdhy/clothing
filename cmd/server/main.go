package main

import (
	"clothing/internal/api"
	"clothing/internal/config"
	"clothing/internal/model"
	"clothing/internal/storage"
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//go:embed web/dist/index.html
var indexHTML string

func main() {
	// 初始化配置
	cfg, err := config.ParseConfig()
	if err != nil {
		logrus.WithError(err).Error("Failed to parse config")
		return
	}

	// 初始化logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	repo, err := model.InitRepository(&cfg)
	if err != nil {
		logrus.WithError(err).Error("failed to initialise repository")
		return
	}

	if repo != nil {
		if err := model.SeedDefaultProviders(context.Background(), repo, cfg); err != nil {
			logrus.WithError(err).Warn("failed to seed default providers")
		}
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		logrus.WithError(err).Error("failed to initialise storage")
		return
	}

	httpHandler, err := api.NewHTTPHandler(cfg, repo, store)
	if err != nil {
		logrus.WithError(err).Error("failed to initialise http handler")
		return
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 添加中间件
	r.Use(LoggingMiddleware())
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	apiGroup := r.Group("/api")

	authGroup := apiGroup.Group("/auth")
	authGroup.GET("/status", httpHandler.AuthStatus)
	authGroup.POST("/register", httpHandler.Register)
	authGroup.POST("/login", httpHandler.Login)
	authGroup.GET("/me", httpHandler.AuthMiddleware(), httpHandler.Me)

	protected := apiGroup.Group("")
	protected.Use(httpHandler.AuthMiddleware())
	protected.GET("/llm/providers", httpHandler.ListProviders)
	protected.POST("/llm", httpHandler.GenerateImage)
	protected.GET("/usage-records", httpHandler.ListUsageRecords)
	protected.GET("/usage-records/:id", httpHandler.GetUsageRecord)
	protected.DELETE("/usage-records/:id", httpHandler.DeleteUsageRecord)

	userAdmin := protected.Group("/users")
	userAdmin.Use(httpHandler.RequireAdmin())
	userAdmin.GET("", httpHandler.ListUsers)
	userAdmin.POST("", httpHandler.CreateUser)
	userAdmin.PATCH(":id", httpHandler.UpdateUser)
	userAdmin.DELETE(":id", httpHandler.DeleteUser)

	providerAdmin := protected.Group("/providers")
	providerAdmin.Use(httpHandler.RequireAdmin())
	providerAdmin.GET("", httpHandler.AdminListProviders)
	providerAdmin.POST("", httpHandler.CreateProvider)
	providerAdmin.GET("/:id", httpHandler.GetProviderDetail)
	providerAdmin.PATCH("/:id", httpHandler.UpdateProvider)
	providerAdmin.DELETE("/:id", httpHandler.DeleteProvider)

	modelAdmin := providerAdmin.Group("/:id/models")
	modelAdmin.GET("", httpHandler.ListProviderModels)
	modelAdmin.POST("", httpHandler.CreateProviderModel)
	modelAdmin.PATCH("/:model_id", httpHandler.UpdateProviderModel)
	modelAdmin.DELETE("/:model_id", httpHandler.DeleteProviderModel)

	if localProvider, ok := store.(storage.LocalBaseDirProvider); ok {
		publicPrefix := strings.TrimSpace(cfg.StoragePublicBaseURL)
		if publicPrefix == "" {
			publicPrefix = "/files"
		}
		if !strings.HasPrefix(publicPrefix, "http://") && !strings.HasPrefix(publicPrefix, "https://") {
			if !strings.HasPrefix(publicPrefix, "/") {
				publicPrefix = "/" + publicPrefix
			}
			r.Static(publicPrefix, localProvider.LocalBaseDir())
		}
	}

	//前端资源
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
	})

	serverHost := fmt.Sprintf("0.0.0.0:%s", cfg.HTTPPort)
	logger.WithField("host", serverHost).Info("服务器启动")
	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:         serverHost,
		Handler:      r,
		ReadTimeout:  900 * time.Second,
		WriteTimeout: 900 * time.Second,
		IdleTimeout:  1200 * time.Second,
	}
	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Error("服务器启动失败")
	}
}

// CORSMiddleware CORS跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// LoggingMiddleware 日志记录中间件
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// 处理请求
		c.Next()
		// 记录请求结束
		duration := time.Since(start)
		logrus.WithFields(logrus.Fields{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"status":    c.Writer.Status(),
			"duration":  duration.String(),
			"size":      c.Writer.Size(),
			"client_ip": c.ClientIP(),
		}).Info("http_request")
	}
}
