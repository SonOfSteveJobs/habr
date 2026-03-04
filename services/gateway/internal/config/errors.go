package config

import "errors"

var (
	ErrHTTPPortNotProvided        = errors.New("GATEWAY_HTTP_PORT is not provided")
	ErrAuthGRPCAddrNotProvided    = errors.New("AUTH_GRPC_ADDR is not provided")
	ErrArticleGRPCAddrNotProvided = errors.New("ARTICLE_GRPC_ADDR is not provided")
	ErrJWTSecretNotProvided       = errors.New("JWT_SECRET is not provided")
	ErrLoggerLevelNotProvided     = errors.New("LOGGER_LEVEL is not provided")
	ErrLoggerAsJsonNotProvided    = errors.New("LOGGER_AS_JSON is not provided")
	ErrLoggerAsJsonInvalid        = errors.New("LOGGER_AS_JSON must be true or false")
)
