package article

import (
	"net/http"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/utils"
)

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request, id gatewayv1.ArticleID) {
	resp, err := h.client.GetArticle(r.Context(), &articlev1.GetArticleRequest{
		Id: id.String(),
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
