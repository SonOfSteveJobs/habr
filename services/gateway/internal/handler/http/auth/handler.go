package auth

import (
	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
)

type Handler struct {
	client authv1.AuthServiceClient
}

func New(client authv1.AuthServiceClient) *Handler {
	return &Handler{client: client}
}
