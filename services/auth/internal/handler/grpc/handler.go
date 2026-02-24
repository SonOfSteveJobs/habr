package authgrpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (uuid.UUID, error)
	Login(ctx context.Context, email, password string) (*model.TokenPair, error)
	RefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) (*model.TokenPair, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error
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

func (h *Handler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	pair, err := h.authService.RefreshToken(ctx, userID, req.GetRefreshToken())
	if err != nil {
		return nil, refreshTokenError(ctx, err)
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := h.authService.Logout(ctx, userID); err != nil {
		return nil, logoutError(ctx, err)
	}

	return &authv1.LogoutResponse{}, nil
}

func (h *Handler) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := h.authService.VerifyEmail(ctx, userID, req.GetCode()); err != nil {
		return nil, verifyEmailError(ctx, err)
	}

	return &authv1.VerifyEmailResponse{}, nil
}
