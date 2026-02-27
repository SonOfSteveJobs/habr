package gatewayhttp

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req gatewayv1.CreateArticleRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.articleClient.CreateArticle(r.Context(), &articlev1.CreateArticleRequest{
		AuthorId: userID.String(),
		Title:    req.Title,
		Content:  req.Content,
	})
	if err != nil {
		handleGRPCError(w, r, err)
		return
	}

	article, err := toArticleResponse(resp.GetArticle())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, article)
}

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request, params gatewayv1.ListArticlesParams) {
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

func toArticleResponse(a *articlev1.Article) (gatewayv1.ArticleResponse, error) {
	if a == nil {
		return gatewayv1.ArticleResponse{}, nil
	}

	id, err := uuid.Parse(a.GetId())
	if err != nil {
		return gatewayv1.ArticleResponse{}, fmt.Errorf("parse article id: %w", err)
	}

	authorID, err := uuid.Parse(a.GetAuthorId())
	if err != nil {
		return gatewayv1.ArticleResponse{}, fmt.Errorf("parse author id: %w", err)
	}

	resp := gatewayv1.ArticleResponse{
		Id:       &id,
		AuthorId: &authorID,
		Title:    new(a.GetTitle()),
		Content:  new(a.GetContent()),
	}

	if a.GetCreatedAt() != nil {
		resp.CreatedAt = new(a.GetCreatedAt().AsTime())
	}
	if a.GetUpdatedAt() != nil {
		resp.UpdatedAt = new(a.GetUpdatedAt().AsTime())
	}

	return resp, nil
}
