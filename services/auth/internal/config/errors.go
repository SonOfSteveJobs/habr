package config

import "errors"

var (
	ErrGRPCPortNotProvided = errors.New("AUTH_GRPC_PORT is not provided")
	ErrDBURINotProvided    = errors.New("DB_URI is not provided")
)
