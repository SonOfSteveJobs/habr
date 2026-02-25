package gatewayhttp

import (
	"net/http"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req gatewayv1.RegisterRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.authClient.Register(r.Context(), &authv1.RegisterRequest{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		handleGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req gatewayv1.LoginRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authClient.Login(r.Context(), &authv1.LoginRequest{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		handleGRPCError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, gatewayv1.TokenPairResponse{
		AccessToken:  &resp.AccessToken,
		RefreshToken: &resp.RefreshToken,
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req gatewayv1.RefreshTokenRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authClient.RefreshToken(r.Context(), &authv1.RefreshTokenRequest{
		UserId:       req.UserId.String(),
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		handleGRPCError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, gatewayv1.TokenPairResponse{
		AccessToken:  &resp.AccessToken,
		RefreshToken: &resp.RefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	_, err := h.authClient.Logout(r.Context(), &authv1.LogoutRequest{
		UserId: userID.String(),
	})
	if err != nil {
		handleGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req gatewayv1.VerifyEmailRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.authClient.VerifyEmail(r.Context(), &authv1.VerifyEmailRequest{
		UserId: req.UserId.String(),
		Code:   req.Code,
	})
	if err != nil {
		handleGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
