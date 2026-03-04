package app

import (
	"github.com/SonOfSteveJobs/habr/services/article/internal/config"
	articlegrpc "github.com/SonOfSteveJobs/habr/services/article/internal/handler/grpc"
	articlerepo "github.com/SonOfSteveJobs/habr/services/article/internal/repository/article"
	cacherepo "github.com/SonOfSteveJobs/habr/services/article/internal/repository/cache"
	"github.com/SonOfSteveJobs/habr/services/article/internal/service"
)

type serviceContainer struct {
	infra *infraContainer

	articleRepo    *articlerepo.Repository
	cacheRepo      *cacherepo.Repository
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

func (c *serviceContainer) CacheRepo() *cacherepo.Repository {
	if c.cacheRepo == nil {
		c.cacheRepo = cacherepo.New(
			c.infra.RedisClient(),
			config.AppConfig().CacheArticlesTTL(),
		)
	}

	return c.cacheRepo
}

func (c *serviceContainer) ArticleService() *service.Service {
	if c.articleService == nil {
		c.articleService = service.New(
			c.ArticleRepo(),
			c.CacheRepo(),
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
