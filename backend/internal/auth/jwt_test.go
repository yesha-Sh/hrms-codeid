package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTokenManagerGenerateAndParseTokens(t *testing.T) {
	manager := NewTokenManager("access-secret", "refresh-secret", 15*time.Minute, 24*time.Hour)
	employeeID := uuid.New()

	accessToken, _, err := manager.GenerateAccessToken("user-1", "manager", &employeeID)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	accessClaims, err := manager.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if accessClaims.UserID != "user-1" {
		t.Fatalf("expected access subject user-1, got %q", accessClaims.UserID)
	}
	if accessClaims.Role != "manager" {
		t.Fatalf("expected role manager, got %q", accessClaims.Role)
	}
	if accessClaims.EmployeeID != employeeID.String() {
		t.Fatalf("expected employee id %q, got %q", employeeID.String(), accessClaims.EmployeeID)
	}

	refreshTokenID, refreshToken, _, err := manager.GenerateRefreshToken("user-1")
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	refreshClaims, err := manager.ParseRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if refreshClaims.UserID != "user-1" {
		t.Fatalf("expected refresh subject user-1, got %q", refreshClaims.UserID)
	}
	if refreshClaims.TokenID != refreshTokenID {
		t.Fatalf("expected refresh token id %q, got %q", refreshTokenID, refreshClaims.TokenID)
	}

	hashA := HashToken(refreshToken)
	hashB := HashToken(refreshToken)
	if hashA == "" || hashA != hashB {
		t.Fatal("expected stable refresh token hashing")
	}
}
