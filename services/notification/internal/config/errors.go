package config

import "errors"

var (
	ErrDBURINotProvided        = errors.New("DB_URI is not provided")
	ErrLoggerLevelNotProvided  = errors.New("LOGGER_LEVEL is not provided")
	ErrLoggerAsJsonNotProvided = errors.New("LOGGER_AS_JSON is not provided")
	ErrLoggerAsJsonInvalid     = errors.New("LOGGER_AS_JSON must be true or false")
	ErrKafkaBrokersNotProvided = errors.New("KAFKA_BROKERS is not provided")
	ErrKafkaTopicNotProvided   = errors.New("KAFKA_TOPIC is not provided")
	ErrKafkaGroupIDNotProvided = errors.New("KAFKA_GROUP_ID is not provided")
)
