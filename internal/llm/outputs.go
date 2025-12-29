package llm

import (
	"clothing/internal/entity"
	"strings"
)

func buildMediaOutputs(urls []string, mediaType string) []entity.MediaOutput {
	if len(urls) == 0 {
		return nil
	}
	normalizedType := strings.ToLower(strings.TrimSpace(mediaType))
	if normalizedType == "" {
		normalizedType = "image"
	}
	outputs := make([]entity.MediaOutput, 0, len(urls))
	for _, url := range urls {
		trimmed := strings.TrimSpace(url)
		if trimmed == "" {
			continue
		}
		outputs = append(outputs, entity.MediaOutput{
			Type: normalizedType,
			URL:  trimmed,
		})
	}
	if len(outputs) == 0 {
		return nil
	}
	return outputs
}
