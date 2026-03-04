package article

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"

	"google.golang.org/grpc"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
)

type mockArticleClient struct {
	createArticleFn func(ctx context.Context, in *articlev1.CreateArticleRequest, opts ...grpc.CallOption) (*articlev1.CreateArticleResponse, error)
	getArticleFn    func(ctx context.Context, in *articlev1.GetArticleRequest, opts ...grpc.CallOption) (*articlev1.GetArticleResponse, error)
	updateArticleFn func(ctx context.Context, in *articlev1.UpdateArticleRequest, opts ...grpc.CallOption) (*articlev1.UpdateArticleResponse, error)
	deleteArticleFn func(ctx context.Context, in *articlev1.DeleteArticleRequest, opts ...grpc.CallOption) (*articlev1.DeleteArticleResponse, error)
	listArticlesFn  func(ctx context.Context, in *articlev1.ListArticlesRequest, opts ...grpc.CallOption) (*articlev1.ListArticlesResponse, error)
}

func (m *mockArticleClient) CreateArticle(ctx context.Context, in *articlev1.CreateArticleRequest, opts ...grpc.CallOption) (*articlev1.CreateArticleResponse, error) {
	return m.createArticleFn(ctx, in, opts...)
}

func (m *mockArticleClient) GetArticle(ctx context.Context, in *articlev1.GetArticleRequest, opts ...grpc.CallOption) (*articlev1.GetArticleResponse, error) {
	return m.getArticleFn(ctx, in, opts...)
}

func (m *mockArticleClient) UpdateArticle(ctx context.Context, in *articlev1.UpdateArticleRequest, opts ...grpc.CallOption) (*articlev1.UpdateArticleResponse, error) {
	return m.updateArticleFn(ctx, in, opts...)
}

func (m *mockArticleClient) DeleteArticle(ctx context.Context, in *articlev1.DeleteArticleRequest, opts ...grpc.CallOption) (*articlev1.DeleteArticleResponse, error) {
	return m.deleteArticleFn(ctx, in, opts...)
}

func (m *mockArticleClient) ListArticles(ctx context.Context, in *articlev1.ListArticlesRequest, opts ...grpc.CallOption) (*articlev1.ListArticlesResponse, error) {
	return m.listArticlesFn(ctx, in, opts...)
}

func newTestHandler(client *mockArticleClient) *Handler {
	return New(client)
}

func makeRequest(method, path, body string) (*httptest.ResponseRecorder, *http.Request) {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	return w, r
}
