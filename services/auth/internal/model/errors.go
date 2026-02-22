package model

import "errors"

var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidPassword     = errors.New("password must contain only letters and digits")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrUserNotFound        = errors.New("user not found")
)
