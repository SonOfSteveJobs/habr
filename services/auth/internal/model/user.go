package model

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

type User struct {
	ID               uuid.UUID
	Email            string
	HashedPassword   string
	IsEmailConfirmed bool
	CreatedAt        time.Time
}

func NewUser(email, password string) (*User, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:             id,
		Email:          email,
		HashedPassword: string(hashedPassword),
	}, nil
}
