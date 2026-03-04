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
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func TestCreateArticle_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	articleID := uuid.Must(uuid.NewV7())

	client := &mockArticleClient{
		createArticleFn: func(_ context.Context, in *articlev1.CreateArticleRequest, _ ...grpc.CallOption) (*articlev1.CreateArticleResponse, error) {
			if in.GetAuthorId() != userID.String() {
				t.Errorf("author_id = %q, want %q", in.GetAuthorId(), userID.String())
			}
			if in.GetTitle() != "Test Title" {
				t.Errorf("title = %q, want %q", in.GetTitle(), "Test Title")
			}

			return &articlev1.CreateArticleResponse{
				Article: &articlev1.Article{
					Id:        articleID.String(),
					AuthorId:  userID.String(),
					Title:     in.GetTitle(),
					Content:   in.GetContent(),
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
			}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/articles", `{"title":"Test Title","content":"Test Content"}`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.CreateArticle(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
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

func TestCreateArticle_NoAuth(t *testing.T) {
	h := newTestHandler(&mockArticleClient{})

	w, r := makeRequest(http.MethodPost, "/api/v1/articles", `{"title":"Test","content":"Test"}`)
	h.CreateArticle(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestCreateArticle_InvalidBody(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	h := newTestHandler(&mockArticleClient{})

	w, r := makeRequest(http.MethodPost, "/api/v1/articles", `{invalid`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.CreateArticle(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateArticle_GRPCError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockArticleClient{
		createArticleFn: func(_ context.Context, _ *articlev1.CreateArticleRequest, _ ...grpc.CallOption) (*articlev1.CreateArticleResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "title is required")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPost, "/api/v1/articles", `{"title":"","content":"Test"}`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.CreateArticle(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
