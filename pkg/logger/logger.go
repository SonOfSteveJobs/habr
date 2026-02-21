package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type ctxKey string

const (
	UserIDKey ctxKey = "user_id"
)

var log zerolog.Logger

func Init(level string, asJSON bool) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	output := os.Stdout

	if asJSON {
		log = zerolog.New(output)
	} else {
		log = zerolog.New(zerolog.ConsoleWriter{Out: output})
	}

	log = log.Level(lvl).With().Timestamp().Logger()
}

func Logger() zerolog.Logger {
	return log
}

func Ctx(ctx context.Context) zerolog.Logger {
	l := log.With()

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		l = l.Str("user_id", userID)
	}

	return l.Logger()
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
