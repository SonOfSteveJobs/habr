package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAccessTokenTTL      = 10 * time.Minute
	defaultRefreshTokenTTL     = 30 * 24 * time.Hour // 30 days
	defaultVerificationCodeTTL = 15 * time.Minute
)

var appConfig *Config

type Config struct {
	grpcPort            string
	dbURI               string
	redisAddr           string
	jwtSecret           string
	accessTokenTTL      time.Duration
	refreshTokenTTL     time.Duration
	verificationCodeTTL time.Duration
	logger              LoggerConfig
	kafka               KafkaConfig
}

func (c *Config) GRPCPort() string                   { return c.grpcPort }
func (c *Config) DBURI() string                      { return c.dbURI }
func (c *Config) RedisAddr() string                  { return c.redisAddr }
func (c *Config) JWTSecret() string                  { return c.jwtSecret }
func (c *Config) AccessTokenTTL() time.Duration      { return c.accessTokenTTL }
func (c *Config) RefreshTokenTTL() time.Duration     { return c.refreshTokenTTL }
func (c *Config) VerificationCodeTTL() time.Duration { return c.verificationCodeTTL }
func (c *Config) Logger() LoggerConfig               { return c.logger }
func (c *Config) Kafka() KafkaConfig                 { return c.kafka }

func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	logger, err := NewLoggerConfig()
	if err != nil {
		return err
	}

	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		return ErrGRPCPortNotProvided
	}

	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		return ErrDBURINotProvided
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return ErrRedisAddrNotProvided
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return ErrJWTSecretNotProvided
	}

	accessTokenTTL := defaultAccessTokenTTL
	if v := os.Getenv("ACCESS_TOKEN_TTL"); v != "" {
		ttl, err := time.ParseDuration(v)
		if nil == err {
			accessTokenTTL = ttl
		}
	}

	refreshTokenTTL := defaultRefreshTokenTTL
	if v := os.Getenv("REFRESH_TOKEN_TTL"); v != "" {
		ttl, err := time.ParseDuration(v)
		if nil == err {
			refreshTokenTTL = ttl
		}
	}

	verificationCodeTTL := defaultVerificationCodeTTL
	if v := os.Getenv("VERIFICATION_CODE_TTL"); v != "" {
		ttl, err := time.ParseDuration(v)
		if nil == err {
			verificationCodeTTL = ttl
		}
	}

	kafka, err := newKafkaConfig()
	if err != nil {
		return err
	}

	appConfig = &Config{
		grpcPort:            grpcPort,
		dbURI:               dbURI,
		redisAddr:           redisAddr,
		jwtSecret:           jwtSecret,
		accessTokenTTL:      accessTokenTTL,
		refreshTokenTTL:     refreshTokenTTL,
		verificationCodeTTL: verificationCodeTTL,
		logger:              logger,
		kafka:               kafka,
	}

	return nil
}

func AppConfig() *Config { return appConfig }
