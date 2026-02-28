package articlegrpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func createArticleError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidTitle):
		return status.Error(codes.InvalidArgument, "invalid title")
	case errors.Is(err, model.ErrInvalidContent):
		return status.Error(codes.InvalidArgument, "invalid content")
	default:
		log := logger.Ctx(ctx)
		log.Error().Err(err).Msg("create article: internal error")

		return status.Error(codes.Internal, "internal error")
	}
}

func listArticlesError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidCursor):
		return status.Error(codes.InvalidArgument, "invalid cursor")
	default:
		log := logger.Ctx(ctx)
		log.Error().Err(err).Msg("list articles: internal error")

		return status.Error(codes.Internal, "internal error")
	}
}
