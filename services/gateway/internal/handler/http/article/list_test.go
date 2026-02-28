package article

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
)

func TestListArticles_Success(t *testing.T) {
	articleID := uuid.Must(uuid.NewV7())
	authorID := uuid.Must(uuid.NewV7())

	client := &mockArticleClient{
		listArticlesFn: func(_ context.Context, _ *articlev1.ListArticlesRequest, _ ...grpc.CallOption) (*articlev1.ListArticlesResponse, error) {
			return &articlev1.ListArticlesResponse{
				Articles: []*articlev1.Article{
					{
						Id:        articleID.String(),
						AuthorId:  authorID.String(),
						Title:     "Test Article",
						Content:   "Test Content",
						CreatedAt: timestamppb.Now(),
						UpdatedAt: timestamppb.Now(),
					},
				},
				NextCursor: "next-cursor-value",
			}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodGet, "/api/v1/articles", "")
	h.ListArticles(w, r, gatewayv1.ListArticlesParams{})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp gatewayv1.ArticleListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Articles == nil || len(*resp.Articles) != 1 {
		t.Fatalf("articles count = %d, want 1", len(*resp.Articles))
	}

	article := (*resp.Articles)[0]
	if article.Id == nil || *article.Id != articleID {
		t.Errorf("article id = %v, want %v", article.Id, articleID)
	}
	if article.Title == nil || *article.Title != "Test Article" {
		t.Errorf("article title = %v, want %q", article.Title, "Test Article")
	}
	if resp.NextCursor == nil || *resp.NextCursor != "next-cursor-value" {
		t.Errorf("next_cursor = %v, want %q", resp.NextCursor, "next-cursor-value")
	}
}

func TestListArticles_WithCursor(t *testing.T) {
	client := &mockArticleClient{
		listArticlesFn: func(_ context.Context, in *articlev1.ListArticlesRequest, _ ...grpc.CallOption) (*articlev1.ListArticlesResponse, error) {
			if in.GetCursor() != "my-cursor" {
				t.Errorf("cursor = %q, want %q", in.GetCursor(), "my-cursor")
			}
			return &articlev1.ListArticlesResponse{}, nil
		},
	}
	h := newTestHandler(client)

	cursor := "my-cursor"
	w, r := makeRequest(http.MethodGet, "/api/v1/articles?cursor=my-cursor", "")
	h.ListArticles(w, r, gatewayv1.ListArticlesParams{Cursor: &cursor})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestListArticles_WithLimit(t *testing.T) {
	client := &mockArticleClient{
		listArticlesFn: func(_ context.Context, in *articlev1.ListArticlesRequest, _ ...grpc.CallOption) (*articlev1.ListArticlesResponse, error) {
			if in.GetLimit() != 10 {
				t.Errorf("limit = %d, want %d", in.GetLimit(), 10)
			}
			return &articlev1.ListArticlesResponse{}, nil
		},
	}
	h := newTestHandler(client)

	limit := 10
	w, r := makeRequest(http.MethodGet, "/api/v1/articles?limit=10", "")
	h.ListArticles(w, r, gatewayv1.ListArticlesParams{Limit: &limit})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestListArticles_EmptyList(t *testing.T) {
	client := &mockArticleClient{
		listArticlesFn: func(_ context.Context, _ *articlev1.ListArticlesRequest, _ ...grpc.CallOption) (*articlev1.ListArticlesResponse, error) {
			return &articlev1.ListArticlesResponse{
				Articles:   nil,
				NextCursor: "",
			}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodGet, "/api/v1/articles", "")
	h.ListArticles(w, r, gatewayv1.ListArticlesParams{})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp gatewayv1.ArticleListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Articles != nil && len(*resp.Articles) != 0 {
		t.Errorf("articles count = %d, want 0", len(*resp.Articles))
	}
	if resp.NextCursor == nil || *resp.NextCursor != "" {
		t.Errorf("next_cursor = %v, want empty string", resp.NextCursor)
	}
}

func TestListArticles_GRPCError(t *testing.T) {
	client := &mockArticleClient{
		listArticlesFn: func(_ context.Context, _ *articlev1.ListArticlesRequest, _ ...grpc.CallOption) (*articlev1.ListArticlesResponse, error) {
			return nil, status.Error(codes.Internal, "database error")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodGet, "/api/v1/articles", "")
	h.ListArticles(w, r, gatewayv1.ListArticlesParams{})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
