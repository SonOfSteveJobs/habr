package config

import (
	"os"

	"github.com/joho/godotenv"
)

var appConfig *Config

type Config struct {
	httpPort        string
	authGRPCAddr    string
	articleGRPCAddr string
	jwtSecret       string
	logger          LoggerConfig
}

func (c *Config) HTTPPort() string        { return c.httpPort }
func (c *Config) AuthGRPCAddr() string    { return c.authGRPCAddr }
func (c *Config) ArticleGRPCAddr() string { return c.articleGRPCAddr }
func (c *Config) JWTSecret() string       { return c.jwtSecret }
func (c *Config) Logger() LoggerConfig    { return c.logger }

func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	logger, err := NewLoggerConfig()
	if err != nil {
		return err
	}

	httpPort := os.Getenv("GATEWAY_HTTP_PORT")
	if httpPort == "" {
		return ErrHTTPPortNotProvided
	}

	authGRPCAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGRPCAddr == "" {
		return ErrAuthGRPCAddrNotProvided
	}

	articleGRPCAddr := os.Getenv("ARTICLE_GRPC_ADDR")
	if articleGRPCAddr == "" {
		return ErrArticleGRPCAddrNotProvided
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return ErrJWTSecretNotProvided
	}

	appConfig = &Config{
		httpPort:        httpPort,
		authGRPCAddr:    authGRPCAddr,
		articleGRPCAddr: articleGRPCAddr,
		jwtSecret:       jwtSecret,
		logger:          logger,
	}

	return nil
}

func AppConfig() *Config { return appConfig }
