package article

import (
	"net/http"

	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/utils"
)

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request, _ gatewayv1.ArticleID) {
	utils.WriteError(w, r, http.StatusNotImplemented, "not implemented")
}

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request, _ gatewayv1.ArticleID) {
	utils.WriteError(w, r, http.StatusNotImplemented, "not implemented")
}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request, _ gatewayv1.ArticleID) {
	utils.WriteError(w, r, http.StatusNotImplemented, "not implemented")
}
