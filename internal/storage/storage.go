package storage

import (
	"clothing/internal/config"
	"context"
	"fmt"
	"strings"
)

const (
	// TypeLocal 表示本地文件系统存储。
	TypeLocal = "local"
	// TypeS3 表示 Amazon S3 或兼容的存储后端。
	TypeS3 = "s3"
	// TypeOSS 表示阿里云 OSS 存储。
	TypeOSS = "oss"
	// TypeCOS 表示腾讯云 COS 存储。
	TypeCOS = "cos"
	// TypeR2 表示 Cloudflare R2 存储。
	TypeR2 = "r2"
)

// SaveOptions 控制存储后端如何持久化文件。
//
// Category 用于在磁盘上组织文件，Extension 提示首选的文件扩展名（不含前导点）。
// 当 Extension 为空时，存储实现应尝试猜测合适的扩展名。
type SaveOptions struct {
	Category     string
	Extension    string
	BaseName     string
	SkipIfExists bool
}

// Storage 是持久化二进制数据并返回存储特定标识符的抽象（例如本地存储的相对路径）。
type Storage interface {
	Save(ctx context.Context, data []byte, opts SaveOptions) (string, error)
}

// LocalBaseDirProvider 由暴露可通过 HTTP 直接提供服务的本地目录的存储驱动实现。
type LocalBaseDirProvider interface {
	LocalBaseDir() string
}

// NewStorage 根据配置实例化存储后端。
func NewStorage(cfg config.Config) (Storage, error) {
	typeName := strings.ToLower(strings.TrimSpace(cfg.StorageType))
	switch typeName {
	case "", TypeLocal:
		return NewLocalStorage(cfg.StorageLocalDir)
	case TypeS3:
		return NewS3Storage(cfg)
	case TypeOSS:
		return NewOSSStorage(cfg)
	case TypeCOS:
		return NewCOSStorage(cfg)
	case TypeR2:
		return NewR2Storage(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
}
