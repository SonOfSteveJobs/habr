package model

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUser_Success(t *testing.T) {
	email := "user@example.com"
	password := "secretpassword"

	user, err := NewUser(email, password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Email != email {
		t.Errorf("email = %q, want %q", user.Email, email)
	}

	if user.ID == uuid.Nil {
		t.Error("ID is nil, want UUID v7")
	}

	if user.ID.Version() != 7 {
		t.Errorf("UUID version = %d, want 7", user.ID.Version())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		t.Error("password hash does not match original password")
	}

	if user.IsEmailConfirmed {
		t.Error("IsEmailConfirmed = true, want false")
	}
}

type testCase struct {
	email string
	name  string
}

func TestNewUser_InvalidEmail(t *testing.T) {
	tests := []testCase{
		{"empty", ""},
		{"no @ sign", "not-an-email"},
		{"no domain", "user@"},
		{"no local part", "@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser(tt.email, "password")
			if !errors.Is(err, ErrInvalidEmail) {
				t.Errorf("error = %v, want ErrInvalidEmail", err)
			}
		})
	}
}

func TestNewUser_InvalidPassword(t *testing.T) {
	passwords := []struct {
		name     string
		password string
	}{
		{"with hyphen", "pass-word"},
		{"with space", "pass word"},
		{"with special chars", "p@ssw0rd!"},
		{"with underscore", "pass_word"},
	}

	for _, tt := range passwords {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser("user@example.com", tt.password)
			if !errors.Is(err, ErrInvalidPassword) {
				t.Errorf("error = %v, want ErrInvalidPassword", err)
			}
		})
	}
}

func TestNewUser_PasswordHashUnique(t *testing.T) {
	u1, err := NewUser("a@example.com", "samepassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u2, err := NewUser("b@example.com", "samepassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u1.HashedPassword == u2.HashedPassword {
		t.Error("two users with same pass hashes")
	}
}

func TestVerifyPassword_Success(t *testing.T) {
	user, err := NewUser("user@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := user.VerifyPassword("correctpassword"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyPassword_Wrong(t *testing.T) {
	user, err := NewUser("user@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = user.VerifyPassword("wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}
