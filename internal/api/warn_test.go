package api_test

import (
	"testing"

	"cdr.dev/coder-doctor/internal/api"
	"cdr.dev/slog/sloggers/slogtest/assert"
)

func TestWarnResult(t *testing.T) {
	t.Parallel()
	res := api.WarnResult("check-name", "something bad happened")

	assert.Equal(t, "name matches", "check-name", res.Name)
	assert.Equal(t, "state matches", api.StateWarning, res.State)
	assert.Equal(t, "summary matches", "something bad happened", res.Summary)
}
