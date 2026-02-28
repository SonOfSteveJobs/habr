package article

import (
	"net/http"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/utils"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request, id gatewayv1.ArticleID) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req gatewayv1.UpdateArticleRequest
	if err := utils.DecodeBody(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateArticle(r.Context(), &articlev1.UpdateArticleRequest{
		Id:       id.String(),
		AuthorId: userID.String(),
		Title:    req.Title,
		Content:  req.Content,
	})
	if err != nil {
		utils.HandleGRPCError(w, r, err)
		return
	}

	article, err := toArticleResponse(resp.GetArticle())
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "internal error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, article)
}
