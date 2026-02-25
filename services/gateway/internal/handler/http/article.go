package gatewayhttp

import (
	"net/http"

	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
)

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request, params gatewayv1.ListArticlesParams) {
	writeError(w, r, http.StatusNotImplemented, "not implemented")
}

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusNotImplemented, "not implemented")
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request, id gatewayv1.ArticleID) {
	writeError(w, r, http.StatusNotImplemented, "not implemented")
}

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request, id gatewayv1.ArticleID) {
	writeError(w, r, http.StatusNotImplemented, "not implemented")
}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request, id gatewayv1.ArticleID) {
	writeError(w, r, http.StatusNotImplemented, "not implemented")
}
