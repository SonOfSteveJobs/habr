package model

import "errors"

var (
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidPassword         = errors.New("password must contain only letters and digits")
	ErrInvalidPasswordSize     = errors.New("password must 1 to 20 sybols")
	ErrInvalidRefreshToken     = errors.New("invalid refresh token")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
)
