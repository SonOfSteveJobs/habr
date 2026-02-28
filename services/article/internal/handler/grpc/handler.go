package articlegrpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

type ArticleService interface {
	CreateArticle(ctx context.Context, authorID uuid.UUID, title, content string) (*model.Article, error)
}

type Handler struct {
	articlev1.UnimplementedArticleServiceServer
	articleService ArticleService
}

func New(articleService ArticleService) *Handler {
	return &Handler{articleService: articleService}
}

func (h *Handler) CreateArticle(ctx context.Context, req *articlev1.CreateArticleRequest) (*articlev1.CreateArticleResponse, error) {
	authorID, err := uuid.Parse(req.GetAuthorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid author_id")
	}

	article, err := h.articleService.CreateArticle(ctx, authorID, req.GetTitle(), req.GetContent())
	if err != nil {
		return nil, createArticleError(ctx, err)
	}

	return &articlev1.CreateArticleResponse{
		Article: &articlev1.Article{
			Id:        article.ID.String(),
			AuthorId:  article.AuthorID.String(),
			Title:     article.Title,
			Content:   article.Content,
			CreatedAt: timestamppb.New(article.CreatedAt),
			UpdatedAt: timestamppb.New(article.UpdatedAt),
		},
	}, nil
}
