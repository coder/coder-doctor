package summarywriter_test

import (
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/summarywriter"
)

func TestSummaryWriter(t *testing.T) {
	t.Parallel()

	w := summarywriter.New(&api.DiscardWriter{})
	for i := 0; i < 5; i++ {
		w.WriteResult(&api.CheckResult{State: api.StatePassed})
		w.WriteResult(&api.CheckResult{State: api.StateInfo})
	}
	for i := 0; i < 3; i++ {
		w.WriteResult(&api.CheckResult{State: api.StateFailed})
		w.WriteResult(&api.CheckResult{State: api.StateSkipped})
	}
	w.WriteResult(&api.CheckResult{State: api.StateWarning})

	expected := summarywriter.SummaryResult{
		Passed:  5,
		Warning: 1,
		Failed:  3,
		Info:    5,
		Skipped: 3,
		Total:   17,
	}
	assert.Equal(t, "summary is correct", expected, w.Summary())
}
