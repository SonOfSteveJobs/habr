package httputil

import (
	"encoding/json"
	"io"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	log := logger.Ctx(r.Context())

	if code >= http.StatusInternalServerError {
		log.Error().Int("status", code).Str("error", msg).Msg("request error")
	} else if code >= http.StatusBadRequest {
		log.Warn().Int("status", code).Str("error", msg).Msg("request error")
	}

	WriteJSON(w, code, gatewayv1.ErrorResponse{Error: &msg})
}

func DecodeBody(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	return json.NewDecoder(r.Body).Decode(v)
}

func grpcToHTTP(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func HandleGRPCError(w http.ResponseWriter, r *http.Request, err error) {
	log := logger.Logger()
	st, ok := status.FromError(err)
	if !ok {
		WriteError(w, r, http.StatusInternalServerError, "internal error")
		return
	}

	WriteError(w, r, grpcToHTTP(st.Code()), st.Message())
	log.Err(err).Msg("handleGRPC Error")
}
