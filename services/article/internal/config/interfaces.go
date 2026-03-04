package config

type LoggerConfig interface {
	Level() string
	AsJson() bool
}
