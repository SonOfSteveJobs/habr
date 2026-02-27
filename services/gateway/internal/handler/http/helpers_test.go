package gatewayhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"

	"google.golang.org/grpc"

	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
)

type mockAuthClient struct {
	registerFn     func(ctx context.Context, in *authv1.RegisterRequest, opts ...grpc.CallOption) (*authv1.RegisterResponse, error)
	loginFn        func(ctx context.Context, in *authv1.LoginRequest, opts ...grpc.CallOption) (*authv1.LoginResponse, error)
	refreshTokenFn func(ctx context.Context, in *authv1.RefreshTokenRequest, opts ...grpc.CallOption) (*authv1.RefreshTokenResponse, error)
	logoutFn       func(ctx context.Context, in *authv1.LogoutRequest, opts ...grpc.CallOption) (*authv1.LogoutResponse, error)
	verifyEmailFn  func(ctx context.Context, in *authv1.VerifyEmailRequest, opts ...grpc.CallOption) (*authv1.VerifyEmailResponse, error)
}

func (m *mockAuthClient) Register(ctx context.Context, in *authv1.RegisterRequest, opts ...grpc.CallOption) (*authv1.RegisterResponse, error) {
	return m.registerFn(ctx, in, opts...)
}

func (m *mockAuthClient) Login(ctx context.Context, in *authv1.LoginRequest, opts ...grpc.CallOption) (*authv1.LoginResponse, error) {
	return m.loginFn(ctx, in, opts...)
}

func (m *mockAuthClient) RefreshToken(ctx context.Context, in *authv1.RefreshTokenRequest, opts ...grpc.CallOption) (*authv1.RefreshTokenResponse, error) {
	return m.refreshTokenFn(ctx, in, opts...)
}

func (m *mockAuthClient) Logout(ctx context.Context, in *authv1.LogoutRequest, opts ...grpc.CallOption) (*authv1.LogoutResponse, error) {
	return m.logoutFn(ctx, in, opts...)
}

func (m *mockAuthClient) VerifyEmail(ctx context.Context, in *authv1.VerifyEmailRequest, opts ...grpc.CallOption) (*authv1.VerifyEmailResponse, error) {
	return m.verifyEmailFn(ctx, in, opts...)
}

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

func newTestHandler(authClient *mockAuthClient, articleClient *mockArticleClient) *Handler {
	return New(authClient, articleClient)
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
