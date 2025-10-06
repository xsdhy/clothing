package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage persists files to the local filesystem.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a LocalStorage instance. The directory is created if
// it does not exist.
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		baseDir = "data/images"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

// LocalBaseDir returns the root directory used for storing files.
func (s *LocalStorage) LocalBaseDir() string {
	return s.baseDir
}

// Save writes the provided bytes to disk and returns a relative path that can
// later be used to build a public URL.
func (s *LocalStorage) Save(ctx context.Context, data []byte, opts SaveOptions) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty payload")
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	relativePath := buildObjectPath(opts.Category, opts.BaseName, opts.Extension)
	absPath := filepath.Join(s.baseDir, filepath.FromSlash(relativePath))
	absDir := filepath.Dir(absPath)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	if opts.SkipIfExists {
		if _, err := os.Stat(absPath); err == nil {
			return relativePath, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat existing file: %w", err)
		}
	}

	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return relativePath, nil
}

var _ Storage = (*LocalStorage)(nil)
var _ LocalBaseDirProvider = (*LocalStorage)(nil)
