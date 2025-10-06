package storage

import (
	"clothing/internal/config"
	"context"
	"fmt"
	"strings"
)

const (
	// TypeLocal represents storage backed by the local filesystem.
	TypeLocal = "local"
	// TypeS3 represents Amazon S3 or compatible storage backends.
	TypeS3 = "s3"
	// TypeOSS represents Alibaba Cloud OSS storage.
	TypeOSS = "oss"
	// TypeCOS represents Tencent Cloud COS storage.
	TypeCOS = "cos"
	// TypeR2 represents Cloudflare R2 storage.
	TypeR2 = "r2"
)

// SaveOptions controls how a file should be persisted by the storage backend.
//
// Category are used to organise files on disk, while Extension
// hints the preferred file extension (without the leading dot).
// When Extension is empty the storage implementation should attempt to guess a
// suitable extension.
type SaveOptions struct {
	Category     string
	Extension    string
	BaseName     string
	SkipIfExists bool
}

// Storage is an abstraction that persists binary blobs and returns a storage
// specific identifier (e.g. a relative path for local storage).
type Storage interface {
	Save(ctx context.Context, data []byte, opts SaveOptions) (string, error)
}

// LocalBaseDirProvider is implemented by storage drivers that expose a local
// directory which can be served directly via HTTP.
type LocalBaseDirProvider interface {
	LocalBaseDir() string
}

// NewStorage instantiates a storage backend based on configuration.
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
