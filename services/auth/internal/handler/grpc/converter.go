package authgrpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func registerError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "invalid email")
	case errors.Is(err, model.ErrInvalidPassword):
		return status.Error(codes.InvalidArgument, "password must contain only letters and digits")
	case errors.Is(err, model.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, "user with this email already exists")
	default:
		log := logger.Ctx(ctx)
		log.Error().Err(err).Msg("register: internal error")

		return status.Error(codes.Internal, "internal error")
	}
}

func loginError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	default:
		log := logger.Ctx(ctx)
		log.Error().Err(err).Msg("login: internal error")

		return status.Error(codes.Internal, "internal error")
	}
}

func refreshTokenError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidRefreshToken):
		return status.Error(codes.Unauthenticated, "invalid refresh token")
	default:
		log := logger.Ctx(ctx)
		log.Error().Err(err).Msg("refresh token: internal error")

		return status.Error(codes.Internal, "internal error")
	}
}

func logoutError(ctx context.Context, err error) error {
	log := logger.Ctx(ctx)
	log.Error().Err(err).Msg("logout: internal error")

	return status.Error(codes.Internal, "internal error")
}

func verifyEmailError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidVerificationCode):
		return status.Error(codes.InvalidArgument, "invalid verification code")
	case errors.Is(err, model.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	default:
		log := logger.Ctx(ctx)
		log.Error().Err(err).Msg("verify email: internal error")

		return status.Error(codes.Internal, "internal error")
	}
}
