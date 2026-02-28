package auth

import (
	"net/http"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/httputil"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req gatewayv1.RegisterRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.client.Register(r.Context(), &authv1.RegisterRequest{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		httputil.HandleGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
