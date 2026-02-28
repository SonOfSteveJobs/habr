package auth

import (
	"net/http"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/httputil"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	_, err := h.client.Logout(r.Context(), &authv1.LogoutRequest{
		UserId: userID.String(),
	})
	if err != nil {
		httputil.HandleGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
