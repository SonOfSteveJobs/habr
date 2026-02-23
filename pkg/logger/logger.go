package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type loggerCtxKey struct{}

var log zerolog.Logger

func Init(level string, asJSON bool) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.CallerFieldName = "caller"
	zerolog.ErrorStackFieldName = "stacktrace"
	zerolog.TimeFieldFormat = time.RFC3339

	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		return strings.ToUpper(l.String())
	}

	output := os.Stdout

	if asJSON {
		log = zerolog.New(output)
	} else {
		log = zerolog.New(zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339})
	}

	log = log.Level(lvl).With().Timestamp().Caller().Logger()
}

func Logger() zerolog.Logger {
	return log
}

func Ctx(ctx context.Context) zerolog.Logger {
	if l, ok := ctx.Value(loggerCtxKey{}).(zerolog.Logger); ok {
		return l
	}

	return log
}
