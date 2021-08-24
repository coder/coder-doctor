package api_test

import (
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

func TestPassResult(t *testing.T) {
	t.Parallel()

	res := api.PassResult("check-name", "check succeeded")

	assert.Equal(t, "name matches", "check-name", res.Name)
	assert.Equal(t, "state matches", api.StatePassed, res.State)
	assert.Equal(t, "summary matches", "check succeeded", res.Summary)
	assert.Equal(t, "error matches", nil, res.Details["error"])
}
