package api_test

import (
	"testing"

	"golang.org/x/xerrors"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

func TestSkippedResult(t *testing.T) {
	t.Parallel()

	checkName := "don't wanna"
	checkSummary := "just because"
	checkErr := xerrors.New("whoops")
	res := api.SkippedResult(checkName, checkSummary, checkErr)
	assert.Equal(t, "name matches", checkName, res.Name)
	assert.Equal(t, "state matches", api.StateSkipped, res.State)
	assert.Equal(t, "summary matches", checkSummary, res.Summary)
	assert.True(t, "details has err", res.Details["error"] != nil)
}
