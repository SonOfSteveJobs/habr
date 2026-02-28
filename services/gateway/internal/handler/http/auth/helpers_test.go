package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"

	"google.golang.org/grpc"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
)

type mockAuthClient struct {
	registerFn     func(ctx context.Context, in *authv1.RegisterRequest, opts ...grpc.CallOption) (*authv1.RegisterResponse, error)
	loginFn        func(ctx context.Context, in *authv1.LoginRequest, opts ...grpc.CallOption) (*authv1.LoginResponse, error)
	refreshTokenFn func(ctx context.Context, in *authv1.RefreshTokenRequest, opts ...grpc.CallOption) (*authv1.RefreshTokenResponse, error)
	logoutFn       func(ctx context.Context, in *authv1.LogoutRequest, opts ...grpc.CallOption) (*authv1.LogoutResponse, error)
	verifyEmailFn  func(ctx context.Context, in *authv1.VerifyEmailRequest, opts ...grpc.CallOption) (*authv1.VerifyEmailResponse, error)
}

func (m *mockAuthClient) Register(ctx context.Context, in *authv1.RegisterRequest, opts ...grpc.CallOption) (*authv1.RegisterResponse, error) {
	return m.registerFn(ctx, in, opts...)
}

func (m *mockAuthClient) Login(ctx context.Context, in *authv1.LoginRequest, opts ...grpc.CallOption) (*authv1.LoginResponse, error) {
	return m.loginFn(ctx, in, opts...)
}

func (m *mockAuthClient) RefreshToken(ctx context.Context, in *authv1.RefreshTokenRequest, opts ...grpc.CallOption) (*authv1.RefreshTokenResponse, error) {
	return m.refreshTokenFn(ctx, in, opts...)
}

func (m *mockAuthClient) Logout(ctx context.Context, in *authv1.LogoutRequest, opts ...grpc.CallOption) (*authv1.LogoutResponse, error) {
	return m.logoutFn(ctx, in, opts...)
}

func (m *mockAuthClient) VerifyEmail(ctx context.Context, in *authv1.VerifyEmailRequest, opts ...grpc.CallOption) (*authv1.VerifyEmailResponse, error) {
	return m.verifyEmailFn(ctx, in, opts...)
}

func newTestHandler(client *mockAuthClient) *Handler {
	return New(client)
}

func makeRequest(method, path, body string) (*httptest.ResponseRecorder, *http.Request) {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	return w, r
}
