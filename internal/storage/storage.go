package storage

import (
	"clothing/internal/config"
	"context"
	"fmt"
)

const (
	// TypeLocal represents storage backed by the local filesystem.
	TypeLocal = "local"
)

// SaveOptions controls how a file should be persisted by the storage backend.
//
// ProviderID and Category are used to organise files on disk, while Extension
// hints the preferred file extension (without the leading dot).
// When Extension is empty the storage implementation should attempt to guess a
// suitable extension.
type SaveOptions struct {
	ProviderID string
	Category   string
	Extension  string
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
	switch cfg.StorageType {
	case "", TypeLocal:
		return NewLocalStorage(cfg.StorageLocalDir)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
}
