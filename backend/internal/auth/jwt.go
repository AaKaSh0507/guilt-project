package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTManager handles JWT token creation and verification
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager creates a new JWTManager with the given secret and TTL
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// Issue creates a new JWT token for the given user and session IDs
func (m *JWTManager) Issue(userID, sessionID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"sid": sessionID,
		"iat": now.Unix(),
		"exp": now.Add(m.ttl).Unix(),
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tkn.SignedString(m.secret)
}

// Verify validates a JWT token and returns the user ID and session ID
func (m *JWTManager) Verify(token string) (string, string, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return "", "", err
	}
	if !parsed.Valid {
		return "", "", ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", ErrInvalidToken
	}

	userID, _ := claims["sub"].(string)
	sessID, _ := claims["sid"].(string)

	if userID == "" || sessID == "" {
		return "", "", ErrInvalidToken
	}

	return userID, sessID, nil
}

// TTL returns the token time-to-live duration
func (m *JWTManager) TTL() time.Duration {
	return m.ttl
}
