package humanwriter_test

import (
	"fmt"
	"strings"
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/humanwriter"
)

var _ = api.ResultWriter(&humanwriter.HumanResultWriter{})
var _ = fmt.Stringer(humanwriter.OutputModeEmoji)

func TestHumanWriter(t *testing.T) {
	t.Parallel()

	var sb strings.Builder

	writer := humanwriter.New(&sb, humanwriter.WithColors(false),
		humanwriter.WithMode(humanwriter.OutputModeText))
	writer.WriteResult(&api.CheckResult{
		Name:    "check-test",
		State:   api.StatePassed,
		Summary: "human writer check test",
	})
	writer.WriteResult(&api.CheckResult{
		State:   api.StateInfo,
		Summary: "summary",
	})

	expected := "PASS human writer check test\n" +
		"INFO summary\n"
	assert.Equal(t, "expected output matches", expected, sb.String())
}
