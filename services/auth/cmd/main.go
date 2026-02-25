package main

import (
	"context"
	"fmt"
	"syscall"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/app"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/config"
)

//goland:noinspection ругается потому что монорепа
func main() {
	if err := config.Load(); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err.Error()))
	}

	cfg := config.AppConfig()

	logger.Init(cfg.Logger().Level(), cfg.Logger().AsJson())
	log := logger.Logger()

	closer.Listen(syscall.SIGINT, syscall.SIGTERM)

	application, err := app.New(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init app")
	}

	application.Run()
}
