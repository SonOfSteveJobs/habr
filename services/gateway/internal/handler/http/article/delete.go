package article

import (
	"net/http"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/utils"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request, id gatewayv1.ArticleID) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	_, err := h.client.DeleteArticle(r.Context(), &articlev1.DeleteArticleRequest{
		Id:       id.String(),
		AuthorId: userID.String(),
	})
	if err != nil {
		utils.HandleGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
