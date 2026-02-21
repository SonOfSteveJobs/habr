package authgrpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Register(_ context.Context, _ *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *Handler) Login(_ context.Context, _ *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
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
