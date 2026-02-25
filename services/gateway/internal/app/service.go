package app

import (
	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
)

type serviceContainer struct {
	infra *infraContainer

	authClient authv1.AuthServiceClient
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
