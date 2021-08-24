package api_test

import (
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

func TestSkippedResult(t *testing.T) {
	t.Parallel()

	checkName := "don't wanna"
	checkSummary := "just because"
	res := api.SkippedResult(checkName, checkSummary)
	assert.Equal(t, "name matches", checkName, res.Name)
	assert.Equal(t, "state matches", api.StateSkipped, res.State)
	assert.Equal(t, "summary matches", checkSummary, res.Summary)
}
