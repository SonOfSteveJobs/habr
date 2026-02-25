package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	jwtPartsLen            = 3
)

type jwtPayload struct {
	UserID string `json:"sub"`
	Exp    int64  `json:"exp"`
}

func Auth(secret string) func(http.Handler) http.Handler {
	secretBytes := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(gatewayv1.BearerScopes) == nil {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			if header == "" {
				writeAuthError(w, r, "missing authorization header")
				return
			}

			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok {
				writeAuthError(w, r, "invalid authorization header")
				return
			}
			log := logger.Ctx(r.Context())
			payload, err := validateJWT(token, secretBytes)
			if err != nil {
				log.Err(err).Msg("validateJWT invalid token")
				writeAuthError(w, r, "invalid token")
				return
			}

			userID, err := uuid.Parse(payload.UserID)
			if err != nil {
				log.Err(err).Msg("userID invalid token")
				writeAuthError(w, r, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func writeAuthError(w http.ResponseWriter, r *http.Request, msg string) {
	log := logger.Ctx(r.Context())
	log.Warn().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("error", msg).
		Msg("auth failed")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func validateJWT(token string, secret []byte) (*jwtPayload, error) {
	parts := strings.Split(token, ".")
	// JWT всегда состоит из трёх частей
	if len(parts) != jwtPartsLen {
		return nil, errInvalidToken
	}

	// берём header + payload, далаем вычисление hmac и сравнение с помощью того же секрета который в auth
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, errInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	// json вида {"user_id":"uuid","exp":1234567890}
	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, errInvalidToken
	}

	if time.Now().Unix() > payload.Exp {
		return nil, errTokenExpired
	}

	return &payload, nil
}

// WithUserID - чисто для тестов
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
