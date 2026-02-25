package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
)

const testSecret = "test-jwt-secret"

func buildJWT(t *testing.T, payload jwtPayload, secret string) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signingInput := header + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s.%s", header, payloadB64, sig)
}

func TestAuth_NoScopes(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth(testSecret)(next)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if !called {
		t.Error("next handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token := buildJWT(t, jwtPayload{
		UserID: userID.String(),
		Exp:    time.Now().Add(10 * time.Minute).Unix(),
	}, testSecret)

	var gotUserID uuid.UUID
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("user ID not found in context")
		}
		gotUserID = id
		w.WriteHeader(http.StatusOK)
	})

	handler := Auth(testSecret)(next)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), gatewayv1.BearerScopes, []string{}))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotUserID != userID {
		t.Errorf("userID = %s, want %s", gotUserID, userID)
	}
}

func TestAuth_MissingHeader(t *testing.T) {
	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), gatewayv1.BearerScopes, []string{}))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_InvalidSignature(t *testing.T) {
	token := buildJWT(t, jwtPayload{
		UserID: uuid.Must(uuid.NewV7()).String(),
		Exp:    time.Now().Add(10 * time.Minute).Unix(),
	}, "wrong-secret")

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), gatewayv1.BearerScopes, []string{}))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	token := buildJWT(t, jwtPayload{
		UserID: uuid.Must(uuid.NewV7()).String(),
		Exp:    time.Now().Add(-1 * time.Minute).Unix(),
	}, testSecret)

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), gatewayv1.BearerScopes, []string{}))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
