package utils

import (
	"fmt"
	"strings"
	"time"
)

func EnsureDataURL(value string) string {
	if strings.HasPrefix(value, "data:") {
		return value
	}
	return "data:image/jpeg;base64," + value
}

func SplitDataURL(value string) (string, string) {
	if !strings.HasPrefix(value, "data:") {
		return "image/jpeg", value
	}

	value = strings.TrimPrefix(value, "data:")
	parts := strings.SplitN(value, ";base64,", 2)
	if len(parts) != 2 {
		return "image/jpeg", ""
	}
	return parts[0], parts[1]
}

func GenerateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
