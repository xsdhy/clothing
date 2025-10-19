package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = bcrypt.DefaultCost

// HashPassword hashes the plain text password.
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

// VerifyPassword compares hashed password with candidate.
func VerifyPassword(hash, candidate string) error {
	if strings.TrimSpace(hash) == "" {
		return errors.New("stored password hash is empty")
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(candidate))
}
