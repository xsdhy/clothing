package storage

import (
	"fmt"
	"mime"
	"path"
	"strings"
	"time"
)

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

func normalizeExtension(ext string) string {
	trimmed := strings.TrimSpace(ext)
	trimmed = strings.TrimPrefix(trimmed, ".")
	if trimmed == "" {
		return "bin"
	}
	return sanitizePathSegment(trimmed)
}

func buildObjectPath(category, baseName, ext string) string {
	now := time.Now().UTC()
	category = sanitizePathSegment(category)
	if category == "" {
		category = "misc"
	}
	normalizedExt := normalizeExtension(ext)
	base := sanitizeFileBase(baseName)
	if base == "" {
		base = fmt.Sprintf("%d", now.UnixNano())
	}
	datedir := fmt.Sprintf("%04d/%02d/%02d", now.Year(), now.Month(), now.Day())
	filename := fmt.Sprintf("%s.%s", base, normalizedExt)
	return path.Join(category, datedir, filename)
}

func detectContentType(ext string) string {
	normalized := normalizeExtension(ext)
	typeName := mime.TypeByExtension("." + normalized)
	if typeName == "" {
		return "application/octet-stream"
	}
	return typeName
}

func joinPrefix(prefix, key string) string {
	cleanPrefix := trimPrefix(prefix)
	if cleanPrefix == "" {
		return strings.TrimLeft(key, "/")
	}
	return path.Join(cleanPrefix, strings.TrimLeft(key, "/"))
}

func trimPrefix(prefix string) string {
	return strings.Trim(strings.TrimSpace(prefix), "/")
}

func sanitizeFileBase(value string) string {
	replaced := strings.ReplaceAll(strings.TrimSpace(value), " ", "-")
	sanitized := sanitizePathSegment(replaced)
	return strings.Trim(sanitized, "-_")
}

// SanitizeToken lowercases the provided token and keeps alphanumeric, dash, and underscore characters only.
func SanitizeToken(value string) string {
	return sanitizePathSegment(value)
}
