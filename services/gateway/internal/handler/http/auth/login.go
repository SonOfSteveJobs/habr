package auth

import (
	"net/http"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/utils"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req gatewayv1.LoginRequest
	if err := utils.DecodeBody(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.Login(r.Context(), &authv1.LoginRequest{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		utils.HandleGRPCError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, gatewayv1.TokenPairResponse{
		AccessToken:  &resp.AccessToken,
		RefreshToken: &resp.RefreshToken,
	})
}
