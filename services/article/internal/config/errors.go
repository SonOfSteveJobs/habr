package config

import "errors"

var (
	ErrGRPCPortNotProvided     = errors.New("ARTICLE_GRPC_PORT is not provided")
	ErrDBURINotProvided        = errors.New("DB_URI is not provided")
	ErrRedisAddrNotProvided    = errors.New("REDIS_ADDR is not provided")
	ErrLoggerLevelNotProvided  = errors.New("LOGGER_LEVEL is not provided")
	ErrLoggerAsJsonNotProvided = errors.New("LOGGER_AS_JSON is not provided")
	ErrLoggerAsJsonInvalid     = errors.New("LOGGER_AS_JSON must be true or false")
	ErrInvalidCacheTTL         = errors.New("CACHE_ARTICLES_TTL is not a valid duration")
)
