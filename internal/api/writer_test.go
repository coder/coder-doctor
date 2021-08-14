package api_test

import (
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

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

func TestCaptureWriter(t *testing.T) {
	t.Parallel()

	w := &api.CaptureWriter{}

	assert.True(t, "initially empty", w.Empty())

	result := &api.CheckResult{
		Name:    "test",
		State:   api.StatePassed,
		Summary: "test result",
	}
	err := w.WriteResult(result)
	assert.Success(t, "captured result", err)
	assert.True(t, "result length 1", w.Len() == 1)
	assert.Equal(t, "get back result", result, w.Get()[0])

	w.Clear()
	assert.True(t, "empty buffer", w.Empty())
}
