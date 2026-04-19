package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type AccessClaims struct {
	UserID     string `json:"sub"`
	Role       string `json:"role"`
	EmployeeID string `json:"employee_id,omitempty"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID  string `json:"sub"`
	TokenID string `json:"jti"`
	jwt.RegisteredClaims
}

func NewTokenManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) TokenManager {
	return TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (m TokenManager) GenerateAccessToken(userID, role string, employeeID *uuid.UUID) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.accessTTL)

	claims := AccessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	if employeeID != nil {
		claims.EmployeeID = employeeID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

func (m TokenManager) GenerateRefreshToken(userID string) (tokenID string, token string, expiresAt time.Time, err error) {
	now := time.Now().UTC()
	expiresAt = now.Add(m.refreshTTL)
	tokenID = uuid.NewString()

	claims := RefreshClaims{
		UserID:  userID,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = jwtToken.SignedString(m.refreshSecret)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return tokenID, token, expiresAt, nil
}

func (m TokenManager) ParseAccessToken(raw string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &AccessClaims{}, func(token *jwt.Token) (any, error) {
		return m.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid access claims")
	}
	return claims, nil
}

func (m TokenManager) ParseRefreshToken(raw string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &RefreshClaims{}, func(token *jwt.Token) (any, error) {
		return m.refreshSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse refresh token: %w", err)
	}
	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh claims")
	}
	return claims, nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
