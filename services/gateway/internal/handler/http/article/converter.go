package article

import (
	"fmt"

	"github.com/google/uuid"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
)

func ptr[T any](v T) *T {
	return &v
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
		Title:    ptr(a.GetTitle()),
		Content:  ptr(a.GetContent()),
	}

	if a.GetCreatedAt() != nil {
		resp.CreatedAt = ptr(a.GetCreatedAt().AsTime())
	}
	if a.GetUpdatedAt() != nil {
		resp.UpdatedAt = ptr(a.GetUpdatedAt().AsTime())
	}

	return resp, nil
}
