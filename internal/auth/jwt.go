package auth

import (
	"clothing/internal/entity"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for authenticated requests.
type Claims struct {
	UserID uint   `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Manager encapsulates JWT generation and validation.
type Manager struct {
	secret []byte
	issuer string
	expiry time.Duration
}

// NewManager creates a new JWT manager.
func NewManager(secret, issuer string, expiry time.Duration) (*Manager, error) {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return nil, errors.New("jwt secret must not be empty")
	}
	if expiry <= 0 {
		expiry = time.Hour * 24
	}
	if strings.TrimSpace(issuer) == "" {
		issuer = "clothing"
	}
	return &Manager{
		secret: []byte(trimmed),
		issuer: issuer,
		expiry: expiry,
	}, nil
}

// GenerateToken issues a signed JWT for the provided user.
func (m *Manager) GenerateToken(user *entity.DbUser) (string, time.Time, error) {
	if m == nil {
		return "", time.Time{}, errors.New("jwt manager is nil")
	}
	if user == nil || user.ID == 0 {
		return "", time.Time{}, errors.New("invalid user for token generation")
	}
	now := time.Now().UTC()
	expiry := now.Add(m.expiry)

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiry, nil
}

// ParseToken validates the token and returns claims.
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	if m == nil {
		return nil, errors.New("jwt manager is nil")
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
