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

func TestUpdateArticle_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	articleID := uuid.Must(uuid.NewV7())
	newTitle := "Updated Title"
	newContent := "Updated Content"

	client := &mockArticleClient{
		updateArticleFn: func(_ context.Context, in *articlev1.UpdateArticleRequest, _ ...grpc.CallOption) (*articlev1.UpdateArticleResponse, error) {
			if in.GetId() != articleID.String() {
				t.Errorf("id = %q, want %q", in.GetId(), articleID.String())
			}
			if in.GetAuthorId() != userID.String() {
				t.Errorf("author_id = %q, want %q", in.GetAuthorId(), userID.String())
			}

			return &articlev1.UpdateArticleResponse{
				Article: &articlev1.Article{
					Id:        articleID.String(),
					AuthorId:  userID.String(),
					Title:     *in.Title,
					Content:   *in.Content,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
			}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPatch, "/api/v1/articles/"+articleID.String(), `{"title":"Updated Title","content":"Updated Content"}`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.UpdateArticle(w, r, articleID)

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
	if resp.Title == nil || *resp.Title != newTitle {
		t.Errorf("title = %v, want %q", resp.Title, newTitle)
	}
	if resp.Content == nil || *resp.Content != newContent {
		t.Errorf("content = %v, want %q", resp.Content, newContent)
	}
}

func TestUpdateArticle_NoAuth(t *testing.T) {
	h := newTestHandler(&mockArticleClient{})

	w, r := makeRequest(http.MethodPatch, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), `{"title":"Test"}`)
	h.UpdateArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUpdateArticle_InvalidBody(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	h := newTestHandler(&mockArticleClient{})

	w, r := makeRequest(http.MethodPatch, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), `{invalid`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.UpdateArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateArticle_NotFound(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockArticleClient{
		updateArticleFn: func(_ context.Context, _ *articlev1.UpdateArticleRequest, _ ...grpc.CallOption) (*articlev1.UpdateArticleResponse, error) {
			return nil, status.Error(codes.NotFound, "article not found")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPatch, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), `{"title":"Test"}`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.UpdateArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUpdateArticle_Forbidden(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockArticleClient{
		updateArticleFn: func(_ context.Context, _ *articlev1.UpdateArticleRequest, _ ...grpc.CallOption) (*articlev1.UpdateArticleResponse, error) {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPatch, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), `{"title":"Test"}`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.UpdateArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestUpdateArticle_GRPCError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockArticleClient{
		updateArticleFn: func(_ context.Context, _ *articlev1.UpdateArticleRequest, _ ...grpc.CallOption) (*articlev1.UpdateArticleResponse, error) {
			return nil, status.Error(codes.Internal, "database error")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodPatch, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), `{"title":"Test"}`)
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.UpdateArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
