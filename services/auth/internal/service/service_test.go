package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

const testJWTSecret = "test-secret"

var (
	testAccessTTL  = 10 * time.Minute
	testRefreshTTL = 30 * 24 * time.Hour
)

type mockUserRepo struct {
	createFn     func(ctx context.Context, user *model.User) error
	getByEmailFn func(ctx context.Context, email string) (*model.User, error)
	createCalled bool
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	m.createCalled = true
	return m.createFn(ctx, user)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.getByEmailFn(ctx, email)
}

type mockTokenRepo struct {
	saveFn     func(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error
	saveCalled bool
}

func (m *mockTokenRepo) Save(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error {
	m.saveCalled = true
	return m.saveFn(ctx, refreshToken, userID, ttl)
}

func newTestService(userRepo *mockUserRepo, tokenRepo *mockTokenRepo) *Service {
	return New(userRepo, tokenRepo, testJWTSecret, testAccessTTL, testRefreshTTL)
}

func testUser(t *testing.T) *model.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	return &model.User{
		ID:             uuid.Must(uuid.NewV7()),
		Email:          "user@example.com",
		HashedPassword: string(hash),
	}
}

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return nil },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	id, err := svc.Register(context.Background(), "user@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id == uuid.Nil {
		t.Error("returned uuid.Nil, want valid UUID")
	}

	if !repo.createCalled {
		t.Error("repo.Create was not called")
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return model.ErrEmailAlreadyExists },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	_, err := svc.Register(context.Background(), "user@example.com", "password")
	if !errors.Is(err, model.ErrEmailAlreadyExists) {
		t.Errorf("error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestRegister_InvalidEmail_RepoNotCalled(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return nil },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	_, err := svc.Register(context.Background(), "invalid", "password")
	if !errors.Is(err, model.ErrInvalidEmail) {
		t.Errorf("error = %v, want ErrInvalidEmail", err)
	}

	if repo.createCalled {
		t.Error("repo.Create was called, want skipped on invalid email")
	}
}

func TestRegister_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return repoErr },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	_, err := svc.Register(context.Background(), "user@example.com", "password")
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}

func TestLogin_Success(t *testing.T) {
	user := testUser(t)

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	tokenRepo := &mockTokenRepo{
		saveFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error { return nil },
	}
	svc := newTestService(userRepo, tokenRepo)

	pair, err := svc.Login(context.Background(), "user@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken is empty")
	}

	if pair.RefreshToken == "" {
		t.Error("RefreshToken is empty")
	}

	if !tokenRepo.saveCalled {
		t.Error("tokenRepo.Save was not called")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, model.ErrUserNotFound
		},
	}
	svc := newTestService(userRepo, &mockTokenRepo{})

	_, err := svc.Login(context.Background(), "noone@example.com", "password")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	user := testUser(t)

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	svc := newTestService(userRepo, &mockTokenRepo{})

	_, err := svc.Login(context.Background(), "user@example.com", "wrongpassword")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_TokenRepoError(t *testing.T) {
	user := testUser(t)
	redisErr := errors.New("redis connection refused")

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	tokenRepo := &mockTokenRepo{
		saveFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error { return redisErr },
	}
	svc := newTestService(userRepo, tokenRepo)

	_, err := svc.Login(context.Background(), "user@example.com", "correctpassword")
	if !errors.Is(err, redisErr) {
		t.Errorf("error = %v, want %v", err, redisErr)
	}
}
