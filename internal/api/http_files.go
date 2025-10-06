package api

import (
	"fmt"
	"strings"
)

func (h *HTTPHandler) publicURL(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	base := h.storagePublicBase
	if base == "" {
		base = "/files"
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(trimmed, "/"))
}
