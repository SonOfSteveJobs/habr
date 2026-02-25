package gatewayhttp

import (
	"encoding/json"
	"io"
	"net/http"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

type Handler struct {
	gatewayv1.Unimplemented
	authClient authv1.AuthServiceClient
}

func New(authClient authv1.AuthServiceClient) *Handler {
	return &Handler{authClient: authClient}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	log := logger.Ctx(r.Context())

	if code >= http.StatusInternalServerError {
		log.Error().Int("status", code).Str("error", msg).Msg("request error")
	} else if code >= http.StatusBadRequest {
		log.Warn().Int("status", code).Str("error", msg).Msg("request error")
	}

	writeJSON(w, code, gatewayv1.ErrorResponse{Error: &msg})
}

func decodeBody(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	return json.NewDecoder(r.Body).Decode(v)
}
