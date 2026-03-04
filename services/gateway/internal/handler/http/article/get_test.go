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

func TestGetArticle_Success(t *testing.T) {
	articleID := uuid.Must(uuid.NewV7())
	authorID := uuid.Must(uuid.NewV7())

	client := &mockArticleClient{
		getArticleFn: func(_ context.Context, in *articlev1.GetArticleRequest, _ ...grpc.CallOption) (*articlev1.GetArticleResponse, error) {
			if in.GetId() != articleID.String() {
				t.Errorf("id = %q, want %q", in.GetId(), articleID.String())
			}

			return &articlev1.GetArticleResponse{
				Article: &articlev1.Article{
					Id:        articleID.String(),
					AuthorId:  authorID.String(),
					Title:     "Test Title",
					Content:   "Test Content",
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
			}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodGet, "/api/v1/articles/"+articleID.String(), "")
	h.GetArticle(w, r, articleID)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp gatewayv1.ArticleResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Id == nil || *resp.Id != articleID {
		t.Errorf("id = %v, want %v", resp.Id, articleID)
	}
	if resp.Title == nil || *resp.Title != "Test Title" {
		t.Errorf("title = %v, want %q", resp.Title, "Test Title")
	}
}

func TestGetArticle_NotFound(t *testing.T) {
	client := &mockArticleClient{
		getArticleFn: func(_ context.Context, _ *articlev1.GetArticleRequest, _ ...grpc.CallOption) (*articlev1.GetArticleResponse, error) {
			return nil, status.Error(codes.NotFound, "article not found")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodGet, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), "")
	h.GetArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetArticle_GRPCError(t *testing.T) {
	client := &mockArticleClient{
		getArticleFn: func(_ context.Context, _ *articlev1.GetArticleRequest, _ ...grpc.CallOption) (*articlev1.GetArticleResponse, error) {
			return nil, status.Error(codes.Internal, "database error")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodGet, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), "")
	h.GetArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
