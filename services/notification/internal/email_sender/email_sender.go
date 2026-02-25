package email_sender

import (
	"context"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/model"
)

type LogSender struct{}

func NewLog() *LogSender {
	return &LogSender{}
}

func (s *LogSender) Send(_ context.Context, event model.UserRegisteredEvent) error {
	log := logger.Logger()
	log.Info().
		Str("event_id", event.EventID).
		Str("email", event.Email).
		Str("code", event.Code).
		Msg("verification email sent")

	return nil
}
