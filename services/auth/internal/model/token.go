package model

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const refreshTokenBytes = 32

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewTokenPair(userID uuid.UUID, secret string, accessTTL time.Duration) (*TokenPair, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
	})

	accessToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	refreshBytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, err
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(refreshBytes)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
