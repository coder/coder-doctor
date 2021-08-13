package api_test

import (
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

var _ = api.ResultWriter(&api.DiscardWriter{})

func TestDiscardWriter(t *testing.T) {
	t.Parallel()

	w := &api.DiscardWriter{}
	err := w.WriteResult(nil)

	assert.Success(t, "discard with nil result", err)

	err = w.WriteResult(&api.CheckResult{
		Name:    "test-check",
		State:   api.StatePassed,
		Summary: "check successful",
	})
	assert.Success(t, "discard with success result", err)
}
