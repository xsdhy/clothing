package main

import (
	"clothing/internal/api"
	"clothing/internal/config"
	_ "embed"
	"fmt"
	"net/http"
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

	httpHandler := api.NewHTTPHandler(cfg)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 添加中间件
	r.Use(LoggingMiddleware())
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.GET("/api/llm/providers", httpHandler.ListProviders)
	r.POST("/api/llm", httpHandler.GenerateImage)

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
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	logger.Info("正在启动HTTP服务器...")
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
