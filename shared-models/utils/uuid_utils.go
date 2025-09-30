package utils

import (
	"strings"

	"github.com/google/uuid"
)

// GenerateUUID 生成UUID，不带-符号
func GenerateUUID() string {
	n := uuid.New().String()
	return strings.ReplaceAll(n, "-", "")
}
