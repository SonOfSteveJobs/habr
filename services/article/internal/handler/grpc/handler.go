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
	ListArticles(ctx context.Context, cursor string, limit int32) (*model.ArticlePage, error)
	GetArticle(ctx context.Context, id uuid.UUID) (*model.Article, error)
	UpdateArticle(ctx context.Context, id, authorID uuid.UUID, title, content *string) (*model.Article, error)
	DeleteArticle(ctx context.Context, id, authorID uuid.UUID) error
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
		Article: toProtoArticle(article),
	}, nil
}

func (h *Handler) ListArticles(ctx context.Context, req *articlev1.ListArticlesRequest) (*articlev1.ListArticlesResponse, error) {
	page, err := h.articleService.ListArticles(ctx, req.GetCursor(), req.GetLimit())
	if err != nil {
		return nil, listArticlesError(ctx, err)
	}

	articles := make([]*articlev1.Article, len(page.Articles))
	for i, a := range page.Articles {
		articles[i] = toProtoArticle(a)
	}

	return &articlev1.ListArticlesResponse{
		Articles:   articles,
		NextCursor: page.NextCursor,
	}, nil
}

func (h *Handler) GetArticle(ctx context.Context, req *articlev1.GetArticleRequest) (*articlev1.GetArticleResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	article, err := h.articleService.GetArticle(ctx, id)
	if err != nil {
		return nil, getArticleError(ctx, err)
	}

	return &articlev1.GetArticleResponse{
		Article: toProtoArticle(article),
	}, nil
}

func (h *Handler) UpdateArticle(ctx context.Context, req *articlev1.UpdateArticleRequest) (*articlev1.UpdateArticleResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	authorID, err := uuid.Parse(req.GetAuthorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid author_id")
	}

	article, err := h.articleService.UpdateArticle(ctx, id, authorID, req.Title, req.Content)
	if err != nil {
		return nil, updateArticleError(ctx, err)
	}

	return &articlev1.UpdateArticleResponse{
		Article: toProtoArticle(article),
	}, nil
}

func (h *Handler) DeleteArticle(ctx context.Context, req *articlev1.DeleteArticleRequest) (*articlev1.DeleteArticleResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	authorID, err := uuid.Parse(req.GetAuthorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid author_id")
	}

	if err := h.articleService.DeleteArticle(ctx, id, authorID); err != nil {
		return nil, deleteArticleError(ctx, err)
	}

	return &articlev1.DeleteArticleResponse{}, nil
}

func toProtoArticle(a *model.Article) *articlev1.Article {
	return &articlev1.Article{
		Id:        a.ID.String(),
		AuthorId:  a.AuthorID.String(),
		Title:     a.Title,
		Content:   a.Content,
		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}
}
