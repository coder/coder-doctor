package api_test

import (
	"errors"
	"testing"

	"cdr.dev/coder-doctor/internal/api"
	"cdr.dev/slog/sloggers/slogtest/assert"
)

func TestErrorResult(t *testing.T) {
	t.Parallel()

	err := errors.New("failed to connect to database")
	res := api.ErrorResult("check-name", "check failed", err)

	assert.Equal(t, "name matches", "check-name", res.Name)
	assert.Equal(t, "state matches", api.StateFailed, res.State)
	assert.Equal(t, "summary matches", "check failed", res.Summary)
	assert.Equal(t, "error matches", err, res.Details["error"])
}
