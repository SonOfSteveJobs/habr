package model

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func NewVerificationCode() (string, error) {
	maxCode := big.NewInt(1_000_000) // 0..999999

	n, err := rand.Int(rand.Reader, maxCode)
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}

	// гарантирует, что код всегда ровно 6 цифр (даже если в начале нули)
	return fmt.Sprintf("%06d", n.Int64()), nil
}
