package app

import (
	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	gatewayhttp "github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/article"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http/auth"
)

type serviceContainer struct {
	infra *infraContainer

	authClient    authv1.AuthServiceClient
	articleClient articlev1.ArticleServiceClient
	handler       *gatewayhttp.Handler
}

func newServiceContainer(infra *infraContainer) *serviceContainer {
	return &serviceContainer{infra: infra}
}

func (c *serviceContainer) AuthClient() authv1.AuthServiceClient {
	if c.authClient == nil {
		c.authClient = authv1.NewAuthServiceClient(c.infra.AuthConn())
	}

	return c.authClient
}

func (c *serviceContainer) ArticleClient() articlev1.ArticleServiceClient {
	if c.articleClient == nil {
		c.articleClient = articlev1.NewArticleServiceClient(c.infra.ArticleConn())
	}

	return c.articleClient
}

func (c *serviceContainer) Handler() *gatewayhttp.Handler {
	if c.handler == nil {
		c.handler = gatewayhttp.New(
			auth.New(c.AuthClient()),
			article.New(c.ArticleClient()),
		)
	}

	return c.handler
}
