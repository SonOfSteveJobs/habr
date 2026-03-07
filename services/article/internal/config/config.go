package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const defaultCacheArticlesTTL = 5 * time.Minute

var appConfig *Config

type Config struct {
	grpcPort         string
	dbURI            string
	redisAddr        string
	logger           LoggerConfig
	cacheArticlesTTL time.Duration
	tracing          *TracingConfig
}

func (c *Config) GRPCPort() string                { return c.grpcPort }
func (c *Config) DBURI() string                   { return c.dbURI }
func (c *Config) RedisAddr() string               { return c.redisAddr }
func (c *Config) Logger() LoggerConfig            { return c.logger }
func (c *Config) CacheArticlesTTL() time.Duration { return c.cacheArticlesTTL }
func (c *Config) Tracing() *TracingConfig         { return c.tracing }

func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	logger, err := NewLoggerConfig()
	if err != nil {
		return err
	}

	grpcPort := os.Getenv("ARTICLE_GRPC_PORT")
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

	cacheArticlesTTL := defaultCacheArticlesTTL
	if v := os.Getenv("CACHE_ARTICLES_TTL"); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return ErrInvalidCacheTTL
		}
		cacheArticlesTTL = parsed
	}

	tracing, err := newTracingConfig()
	if err != nil {
		return err
	}

	appConfig = &Config{
		grpcPort:         grpcPort,
		dbURI:            dbURI,
		redisAddr:        redisAddr,
		logger:           logger,
		cacheArticlesTTL: cacheArticlesTTL,
		tracing:          tracing,
	}

	return nil
}

func AppConfig() *Config { return appConfig }
