package article

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

func TestDeleteArticle_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	articleID := uuid.Must(uuid.NewV7())

	client := &mockArticleClient{
		deleteArticleFn: func(_ context.Context, in *articlev1.DeleteArticleRequest, _ ...grpc.CallOption) (*articlev1.DeleteArticleResponse, error) {
			if in.GetId() != articleID.String() {
				t.Errorf("id = %q, want %q", in.GetId(), articleID.String())
			}
			if in.GetAuthorId() != userID.String() {
				t.Errorf("author_id = %q, want %q", in.GetAuthorId(), userID.String())
			}

			return &articlev1.DeleteArticleResponse{}, nil
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodDelete, "/api/v1/articles/"+articleID.String(), "")
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.DeleteArticle(w, r, articleID)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestDeleteArticle_NoAuth(t *testing.T) {
	h := newTestHandler(&mockArticleClient{})

	w, r := makeRequest(http.MethodDelete, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), "")
	h.DeleteArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestDeleteArticle_NotFound(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockArticleClient{
		deleteArticleFn: func(_ context.Context, _ *articlev1.DeleteArticleRequest, _ ...grpc.CallOption) (*articlev1.DeleteArticleResponse, error) {
			return nil, status.Error(codes.NotFound, "article not found")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodDelete, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), "")
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.DeleteArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDeleteArticle_Forbidden(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockArticleClient{
		deleteArticleFn: func(_ context.Context, _ *articlev1.DeleteArticleRequest, _ ...grpc.CallOption) (*articlev1.DeleteArticleResponse, error) {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodDelete, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), "")
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.DeleteArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestDeleteArticle_GRPCError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	client := &mockArticleClient{
		deleteArticleFn: func(_ context.Context, _ *articlev1.DeleteArticleRequest, _ ...grpc.CallOption) (*articlev1.DeleteArticleResponse, error) {
			return nil, status.Error(codes.Internal, "database error")
		},
	}
	h := newTestHandler(client)

	w, r := makeRequest(http.MethodDelete, "/api/v1/articles/"+uuid.Must(uuid.NewV7()).String(), "")
	ctx := middleware.WithUserID(r.Context(), userID)
	r = r.WithContext(ctx)

	h.DeleteArticle(w, r, uuid.Must(uuid.NewV7()))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
