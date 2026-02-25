package app

type serviceContainer struct {
	infra *infraContainer
}

func newServiceContainer(infra *infraContainer) *serviceContainer {
	return &serviceContainer{infra: infra}
}
