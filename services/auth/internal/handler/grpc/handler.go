package authgrpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (uuid.UUID, error)
	Login(ctx context.Context, email, password string) (*model.TokenPair, error)
}

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	authService AuthService
}

func New(authService AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	_, err := h.authService.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, registerError(ctx, err)
	}

	return &authv1.RegisterResponse{}, nil
}

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	pair, err := h.authService.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, loginError(ctx, err)
	}

	return &authv1.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

func (h *Handler) RefreshToken(_ context.Context, _ *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *Handler) VerifyEmail(_ context.Context, _ *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *Handler) Logout(_ context.Context, _ *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

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
