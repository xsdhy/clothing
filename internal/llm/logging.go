package llm

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

const logSnippetLimit = 120

func providerLogger(ctx context.Context, providerID, model string) *logrus.Entry {
	fields := logrus.Fields{
		"provider": providerID,
	}
	if trimmedModel := strings.TrimSpace(model); trimmedModel != "" {
		fields["model"] = trimmedModel
	}

	entry := logrus.WithFields(fields)
	if ctx != nil {
		entry = entry.WithContext(ctx)
	}
	return entry
}

func logSnippet(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= logSnippetLimit {
		return value
	}

	return string(runes[:logSnippetLimit]) + "..."
}
