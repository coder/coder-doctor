package kube

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
)

// newTestHTTPServer creates a HTTP server that just returns the given status code and response.
func newTestHTTPServer(t *testing.T, statusCode int, resp interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		err := json.NewEncoder(w).Encode(resp)
		assert.Success(t, "failed to encode response", err)
	}))
	return srv
}
