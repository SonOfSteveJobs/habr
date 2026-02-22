package model

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestNewTokenPair_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	secret := "test-secret-key"
	accessTTL := 10 * time.Minute

	pair, err := NewTokenPair(userID, secret, accessTTL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken is empty")
	}

	if pair.RefreshToken == "" {
		t.Error("RefreshToken is empty")
	}

	// refresh token should be valid base64url
	decoded, err := base64.URLEncoding.DecodeString(pair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken is not valid base64url: %v", err)
	}

	if len(decoded) != refreshTokenBytes {
		t.Errorf("RefreshToken decoded length = %d, want %d", len(decoded), refreshTokenBytes)
	}
}

func TestNewTokenPair_JWTClaims(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	secret := "test-secret-key"
	accessTTL := 10 * time.Minute

	pair, err := NewTokenPair(userID, secret, accessTTL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := jwt.Parse(pair.AccessToken, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		t.Fatalf("failed to get subject: %v", err)
	}

	if sub != userID.String() {
		t.Errorf("subject = %q, want %q", sub, userID.String())
	}

	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		t.Fatalf("failed to get expiration: %v", err)
	}

	expectedExp := time.Now().Add(accessTTL)
	diff := exp.Sub(expectedExp)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expiration diff = %v, want within 1 second", diff)
	}
}

func TestNewTokenPair_UniqueRefreshTokens(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	p1, err := NewTokenPair(userID, "secret", 10*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p2, err := NewTokenPair(userID, "secret", 10*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p1.RefreshToken == p2.RefreshToken {
		t.Error("two calls produced identical refresh tokens")
	}
}
