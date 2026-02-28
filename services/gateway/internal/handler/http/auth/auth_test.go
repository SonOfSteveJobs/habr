package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func TestRegister_Success(t *testing.T) {
	client := &mockAuthClient{
		registerFn: func(_ context.Context, _ *authv1.RegisterRequest, _ ...grpc.CallOption) (*authv1.RegisterResponse, error) {
			return &authv1.RegisterResponse{}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/register", `{"email":"user@example.com","password":"pass123"}`)
	h.Register(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestRegister_InvalidBody(t *testing.T) {
	h := newTestHandler(&mockAuthClient{})

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/register", `{invalid`)
	h.Register(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegister_GRPCAlreadyExists(t *testing.T) {
	client := &mockAuthClient{
		registerFn: func(_ context.Context, _ *authv1.RegisterRequest, _ ...grpc.CallOption) (*authv1.RegisterResponse, error) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/register", `{"email":"user@example.com","password":"pass123"}`)
	h.Register(w, r)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestLogin_Success(t *testing.T) {
	client := &mockAuthClient{
		loginFn: func(_ context.Context, _ *authv1.LoginRequest, _ ...grpc.CallOption) (*authv1.LoginResponse, error) {
			return &authv1.LoginResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/login", `{"email":"user@example.com","password":"pass123"}`)
	h.Login(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp gatewayv1.TokenPairResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.AccessToken == nil || *resp.AccessToken != "access-token" {
		t.Errorf("access_token = %v, want %q", resp.AccessToken, "access-token")
	}
	if resp.RefreshToken == nil || *resp.RefreshToken != "refresh-token" {
		t.Errorf("refresh_token = %v, want %q", resp.RefreshToken, "refresh-token")
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	h := newTestHandler(&mockAuthClient{})

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/login", `{invalid`)
	h.Login(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLogin_GRPCUnauthenticated(t *testing.T) {
	client := &mockAuthClient{
		loginFn: func(_ context.Context, _ *authv1.LoginRequest, _ ...grpc.CallOption) (*authv1.LoginResponse, error) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/login", `{"email":"user@example.com","password":"wrong"}`)
	h.Login(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockAuthClient{
		refreshTokenFn: func(_ context.Context, _ *authv1.RefreshTokenRequest, _ ...grpc.CallOption) (*authv1.RefreshTokenResponse, error) {
			return &authv1.RefreshTokenResponse{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
			}, nil
		},
	}
	h := newTestHandler(client)

	body := `{"user_id":"` + userID.String() + `","refresh_token":"old-refresh"}`
	w, r := makeRequest(http.MethodPost, "/api/v1/auth/refresh", body)
	h.RefreshToken(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp gatewayv1.TokenPairResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.AccessToken == nil || *resp.AccessToken != "new-access" {
		t.Errorf("access_token = %v, want %q", resp.AccessToken, "new-access")
	}
}

func TestRefreshToken_GRPCError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockAuthClient{
		refreshTokenFn: func(_ context.Context, _ *authv1.RefreshTokenRequest, _ ...grpc.CallOption) (*authv1.RefreshTokenResponse, error) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		},
	}
	h := newTestHandler(client)

	body := `{"user_id":"` + userID.String() + `","refresh_token":"bad-token"}`
	w, r := makeRequest(http.MethodPost, "/api/v1/auth/refresh", body)
	h.RefreshToken(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLogout_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockAuthClient{
		logoutFn: func(_ context.Context, _ *authv1.LogoutRequest, _ ...grpc.CallOption) (*authv1.LogoutResponse, error) {
			return &authv1.LogoutResponse{}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/logout", "")
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.Logout(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestLogout_NoUserID(t *testing.T) {
	h := newTestHandler(&mockAuthClient{})

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/logout", "")
	h.Logout(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLogout_GRPCError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockAuthClient{
		logoutFn: func(_ context.Context, _ *authv1.LogoutRequest, _ ...grpc.CallOption) (*authv1.LogoutResponse, error) {
			return nil, status.Error(codes.Internal, "redis error")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/auth/logout", "")
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.Logout(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestVerifyEmail_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockAuthClient{
		verifyEmailFn: func(_ context.Context, _ *authv1.VerifyEmailRequest, _ ...grpc.CallOption) (*authv1.VerifyEmailResponse, error) {
			return &authv1.VerifyEmailResponse{}, nil
		},
	}
	h := newTestHandler(client)

	body := `{"user_id":"` + userID.String() + `","code":"123456"}`
	w, r := makeRequest(http.MethodPost, "/api/v1/auth/verify-email", body)
	h.VerifyEmail(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestVerifyEmail_GRPCError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockAuthClient{
		verifyEmailFn: func(_ context.Context, _ *authv1.VerifyEmailRequest, _ ...grpc.CallOption) (*authv1.VerifyEmailResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "invalid code")
		},
	}
	h := newTestHandler(client)

	body := `{"user_id":"` + userID.String() + `","code":"000000"}`
	w, r := makeRequest(http.MethodPost, "/api/v1/auth/verify-email", body)
	h.VerifyEmail(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
