package article

import (
	"net/http"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/httputil"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req gatewayv1.CreateArticleRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateArticle(r.Context(), &articlev1.CreateArticleRequest{
		AuthorId: userID.String(),
		Title:    req.Title,
		Content:  req.Content,
	})
	if err != nil {
		httputil.HandleGRPCError(w, r, err)
		return
	}

	article, err := toArticleResponse(resp.GetArticle())
	if err != nil {
		httputil.WriteError(w, r, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, article)
}
