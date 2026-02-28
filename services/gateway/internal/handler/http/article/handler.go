package article

import (
	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
)

type Handler struct {
	client articlev1.ArticleServiceClient
}

func New(client articlev1.ArticleServiceClient) *Handler {
	return &Handler{client: client}
}
