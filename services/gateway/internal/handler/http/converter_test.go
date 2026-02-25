package gatewayhttp

import (
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestGrpcToHTTP(t *testing.T) {
	tests := []struct {
		name     string
		code     codes.Code
		expected int
	}{
		{"InvalidArgument", codes.InvalidArgument, http.StatusBadRequest},
		{"Unauthenticated", codes.Unauthenticated, http.StatusUnauthorized},
		{"PermissionDenied", codes.PermissionDenied, http.StatusForbidden},
		{"NotFound", codes.NotFound, http.StatusNotFound},
		{"AlreadyExists", codes.AlreadyExists, http.StatusConflict},
		{"Internal", codes.Internal, http.StatusInternalServerError},
		{"Unknown", codes.Unknown, http.StatusInternalServerError},
		{"Unavailable", codes.Unavailable, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcToHTTP(tt.code)
			if got != tt.expected {
				t.Errorf("grpcToHTTP(%v) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}
