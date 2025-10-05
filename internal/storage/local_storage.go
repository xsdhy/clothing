package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
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

	category := sanitizePathSegment(opts.Category)
	if category == "" {
		category = "misc"
	}

	ext := strings.TrimPrefix(strings.TrimSpace(opts.Extension), ".")
	if ext == "" {
		ext = "bin"
	}

	now := time.Now().UTC()
	datedir := fmt.Sprintf("%04d/%02d/%02d", now.Year(), now.Month(), now.Day())
	relativeDir := path.Join(category, datedir)
	filename := fmt.Sprintf("%d.%s", now.UnixNano(), ext)
	relativePath := path.Join(relativeDir, filename)

	absDir := filepath.Join(s.baseDir, filepath.FromSlash(relativeDir))
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	absPath := filepath.Join(s.baseDir, filepath.FromSlash(relativePath))
	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return relativePath, nil
}

func sanitizePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	builder := strings.Builder{}
	builder.Grow(len(value))
	for i := 0; i < len(value); i++ {
		ch := value[i]
		switch {
		case ch >= 'a' && ch <= 'z', ch >= '0' && ch <= '9':
			builder.WriteByte(ch)
		case ch >= 'A' && ch <= 'Z':
			builder.WriteByte(ch + 32)
		case ch == '-', ch == '_':
			builder.WriteByte(ch)
		}
	}
	return builder.String()
}

var _ Storage = (*LocalStorage)(nil)
var _ LocalBaseDirProvider = (*LocalStorage)(nil)
