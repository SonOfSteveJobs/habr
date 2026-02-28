package gatewayhttp

import (
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/article"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/auth"
)

type (
	ArticleHandler = article.Handler
	AuthHandler    = auth.Handler
)

type Handler struct {
	*ArticleHandler
	*AuthHandler
}

func New(authHandler *auth.Handler, articleHandler *article.Handler) *Handler {
	return &Handler{
		ArticleHandler: articleHandler,
		AuthHandler:    authHandler,
	}
}
