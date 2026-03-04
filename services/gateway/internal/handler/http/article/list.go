package article

import (
	"net/http"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/utils"
)

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request, params gatewayv1.ListArticlesParams) {
	req := &articlev1.ListArticlesRequest{}
	if params.Cursor != nil {
		req.Cursor = *params.Cursor
	}
	if params.Limit != nil {
		req.Limit = int32(*params.Limit)
	}

	resp, err := h.client.ListArticles(r.Context(), req)
	if err != nil {
		utils.HandleGRPCError(w, r, err)
		return
	}

	articles := make([]gatewayv1.ArticleResponse, 0, len(resp.GetArticles()))
	for _, a := range resp.GetArticles() {
		article, err := toArticleResponse(a)
		if err != nil {
			utils.WriteError(w, r, http.StatusInternalServerError, "internal error")
			return
		}
		articles = append(articles, article)
	}

	utils.WriteJSON(w, http.StatusOK, gatewayv1.ArticleListResponse{
		Articles:   &articles,
		NextCursor: new(resp.GetNextCursor()),
	})
}
