package app

import (
	articlegrpc "github.com/SonOfSteveJobs/habr/services/article/internal/handler/grpc"
	articlerepo "github.com/SonOfSteveJobs/habr/services/article/internal/repository/article"
	"github.com/SonOfSteveJobs/habr/services/article/internal/service"
)

type serviceContainer struct {
	infra *infraContainer

	articleRepo    *articlerepo.Repository
	articleService *service.Service
	handler        *articlegrpc.Handler
}

func newServiceContainer(infra *infraContainer) *serviceContainer {
	return &serviceContainer{infra: infra}
}

func (c *serviceContainer) ArticleRepo() *articlerepo.Repository {
	if c.articleRepo == nil {
		c.articleRepo = articlerepo.New(c.infra.TxManager())
	}

	return c.articleRepo
}

func (c *serviceContainer) ArticleService() *service.Service {
	if c.articleService == nil {
		c.articleService = service.New(
			c.ArticleRepo(),
			c.infra.TxManager(),
		)
	}

	return c.articleService
}

func (c *serviceContainer) Handler() *articlegrpc.Handler {
	if c.handler == nil {
		c.handler = articlegrpc.New(c.ArticleService())
	}

	return c.handler
}
