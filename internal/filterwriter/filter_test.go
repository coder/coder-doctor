package filterwriter_test

import (
	"strings"
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/filterwriter"
	"github.com/cdr/coder-doctor/internal/humanwriter"
)

func TestFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name           string
		Options        []filterwriter.Option
		Results        []*api.CheckResult
		ExpectedOutput string
	}{
		{
			// Default behavior, only logging items that would be filtered
			Name:    "default-filtered",
			Options: nil,
			Results: []*api.CheckResult{
				{
					State:   api.StateInfo,
					Summary: "filtered message",
				}, {
					State:   api.StateSkipped,
					Summary: "skipped message",
				}, {
					State:   api.StatePassed,
					Summary: "passed result",
				}, {
					State:   api.StateFailed,
					Summary: "failed result",
				}, {
					State:   api.StateInfo,
					Summary: "info message",
				},
			},
			ExpectedOutput: "PASS passed result\nFAIL failed result\n",
		}, {
			// Everything filtered out, write at each level and verify the
			// result is empty
			Name: "filter-everything",
			Options: []filterwriter.Option{
				filterwriter.WithFilterState(api.StatePassed),
				filterwriter.WithFilterState(api.StateWarning),
				filterwriter.WithFilterState(api.StateFailed),
				filterwriter.WithFilterState(api.StateInfo),
				filterwriter.WithFilterState(api.StateSkipped),
			},
			Results: []*api.CheckResult{
				{
					State:   api.StateFailed,
					Summary: "failed",
				},
			},
			ExpectedOutput: "",
		}, {
			// Accept informational messages only
			Name: "accept-informational",
			Options: []filterwriter.Option{
				filterwriter.WithFilterState(api.StatePassed),
				filterwriter.WithFilterState(api.StateWarning),
				filterwriter.WithFilterState(api.StateFailed),
				filterwriter.WithFilterState(api.StateInfo),
				filterwriter.WithFilterState(api.StateSkipped),
				filterwriter.WithAcceptState(api.StateInfo),
			},
			Results: []*api.CheckResult{
				{
					State:   api.StateInfo,
					Summary: "info message",
				}, {
					State:   api.StateSkipped,
					Summary: "skipped message",
				}, {
					State:   api.StatePassed,
					Summary: "passed result",
				}, {
					State:   api.StateFailed,
					Summary: "failed result",
				}, {
					State:   api.StateInfo,
					Summary: "info result",
				},
			},
			ExpectedOutput: "INFO info message\nINFO info result\n",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			var sb strings.Builder

			w := filterwriter.Must(humanwriter.New(&sb), test.Options...)
			for _, result := range test.Results {
				w.WriteResult(result)
			}

			assert.Equal(t, test.Name, test.ExpectedOutput, sb.String())
		})
	}
}
