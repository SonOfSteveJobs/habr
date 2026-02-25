package app

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/config"
)

type infraContainer struct {
	authConn    *grpc.ClientConn
	articleConn *grpc.ClientConn
}

func newInfraContainer() (*infraContainer, error) {
	c := &infraContainer{}

	if err := c.initAuthConn(); err != nil {
		return nil, fmt.Errorf("auth grpc conn: %w", err)
	}

	if err := c.initArticleConn(); err != nil {
		return nil, fmt.Errorf("article grpc conn: %w", err)
	}

	return c, nil
}

func (c *infraContainer) AuthConn() *grpc.ClientConn    { return c.authConn }
func (c *infraContainer) ArticleConn() *grpc.ClientConn { return c.articleConn }

func (c *infraContainer) initAuthConn() error {
	conn, err := grpc.NewClient(
		config.AppConfig().AuthGRPCAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	closer.AddNamed("auth grpc conn", func(_ context.Context) error {
		return conn.Close()
	})

	c.authConn = conn
	return nil
}

func (c *infraContainer) initArticleConn() error {
	conn, err := grpc.NewClient(
		config.AppConfig().ArticleGRPCAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	closer.AddNamed("article grpc conn", func(_ context.Context) error {
		return conn.Close()
	})

	c.articleConn = conn
	return nil
}
