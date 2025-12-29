package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = bcrypt.DefaultCost

// HashPassword 对明文密码进行哈希处理
func HashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("password must not be empty")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), defaultBcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword 验证密码是否与存储的哈希值匹配
func VerifyPassword(hash, candidate string) error {
	if strings.TrimSpace(hash) == "" {
		return errors.New("stored password hash is empty")
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(candidate))
}
