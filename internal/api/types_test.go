package api_test

import (
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

func TestKnownStates(t *testing.T) {
	t.Parallel()

	states := []api.CheckState{
		api.StatePassed,
		api.StateWarning,
		api.StateFailed,
		api.StateInfo,
		api.StateSkipped,
	}

	for _, state := range states {
		state := state

		t.Run(state.String(), func(t *testing.T) {
			t.Parallel()

			emoji, err := state.Emoji()
			assert.Success(t, "state.Emoji() error non-nil", err)
			assert.True(t, "state.Emoji() is non-empty", len(emoji) > 0)
			_ = state.MustEmoji()

			text, err := state.Text()
			assert.Success(t, "state.Text() error non-nil", err)
			assert.True(t, "state.Text() is non-empty", len(text) > 0)
			_ = state.MustText()

			colorFunc, err := state.Color()
			assert.Success(t, "state.Color() error non-nil", err)
			assert.True(t, "state.Color() is non-nil", colorFunc != nil)
			_ = state.MustColor()

			str := state.String()
			assert.True(t, "state.String() is non-empty", len(str) > 0)
		})
	}
}
