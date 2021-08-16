package api_test

import (
	"errors"
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
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
