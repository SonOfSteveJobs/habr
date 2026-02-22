package model

import (
	"net/mail"
	"time"
	"unicode"

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

	if err := validatePassword(password); err != nil {
		return nil, err
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

func validatePassword(password string) error {
	for _, r := range password {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return ErrInvalidPassword
		}
	}

	return nil
}

func (u *User) VerifyPassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}
