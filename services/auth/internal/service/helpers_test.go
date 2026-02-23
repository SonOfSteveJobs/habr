package service

import (
	"context"
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
	saveFn       func(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error
	validateFn   func(ctx context.Context, refreshToken string, userID uuid.UUID) error
	deleteFn     func(ctx context.Context, userID uuid.UUID) error
	saveCalled   bool
	deleteCalled bool
}

func (m *mockTokenRepo) Save(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error {
	m.saveCalled = true
	return m.saveFn(ctx, refreshToken, userID, ttl)
}

func (m *mockTokenRepo) Validate(ctx context.Context, refreshToken string, userID uuid.UUID) error {
	return m.validateFn(ctx, refreshToken, userID)
}

func (m *mockTokenRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	m.deleteCalled = true
	return m.deleteFn(ctx, userID)
}

type mockOutboxRepo struct {
	insertFn func(ctx context.Context, event model.OutboxEvent) error
}

func (m *mockOutboxRepo) Insert(ctx context.Context, event model.OutboxEvent) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, event)
	}
	return nil
}

type mockTxManager struct{}

func (m *mockTxManager) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newTestService(userRepo *mockUserRepo, tokenRepo *mockTokenRepo) *Service {
	return New(userRepo, tokenRepo, &mockOutboxRepo{}, &mockTxManager{}, testJWTSecret, "test-topic", testAccessTTL, testRefreshTTL)
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
