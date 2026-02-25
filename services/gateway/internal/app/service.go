package app

import (
	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	gatewayhttp "github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/http"
)

type serviceContainer struct {
	infra *infraContainer

	authClient authv1.AuthServiceClient
	handler    *gatewayhttp.Handler
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

func (c *serviceContainer) Handler() *gatewayhttp.Handler {
	if c.handler == nil {
		c.handler = gatewayhttp.New(c.AuthClient())
	}

	return c.handler
}
