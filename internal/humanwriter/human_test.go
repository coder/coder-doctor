package humanwriter_test

import (
	"strings"
	"testing"

	"cdr.dev/coder-doctor/internal/api"
	"cdr.dev/coder-doctor/internal/humanwriter"
	"cdr.dev/slog/sloggers/slogtest/assert"
)

func TestHumanWriter(t *testing.T) {
	t.Parallel()

	modes := []humanwriter.OutputMode{
		humanwriter.OutputModeEmoji,
		humanwriter.OutputModeText,
	}

	for _, mode := range modes {
		mode := mode
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()

			var sb strings.Builder

			writer := humanwriter.New(&sb,
				humanwriter.WithColors(false),
				humanwriter.WithMode(mode))
			writer.WriteResult(&api.CheckResult{
				Name:    "check-test",
				State:   api.StatePassed,
				Summary: "human writer check test",
			})
			writer.WriteResult(&api.CheckResult{
				State:   api.StateInfo,
				Summary: "summary",
			})
			writer.WriteResult(&api.CheckResult{
				State:   api.StateFailed,
				Summary: "failed message",
			})
			writer.WriteResult(&api.CheckResult{
				State:   api.StateWarning,
				Summary: "",
			})
			writer.WriteResult(&api.CheckResult{
				State:   api.StateSkipped,
				Summary: "skipped check",
			})

			var expected string
			switch mode {
			case humanwriter.OutputModeEmoji:
				expected = "✓ human writer check test\n" +
					"🔔 summary\n" +
					"✗ failed message\n" +
					"⚠️ \n" +
					"⏩ skipped check\n"
			case humanwriter.OutputModeText:
				expected = "PASS human writer check test\n" +
					"INFO summary\n" +
					"FAIL failed message\n" +
					"WARN \n" +
					"SKIP skipped check\n"
			}

			assert.Equal(t, "expected output matches", expected, sb.String())
		})
	}
}
