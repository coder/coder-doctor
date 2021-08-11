package api

import (
	"fmt"
	"io"
)

func ErrorResult(name string, summary string, err error) *CheckResult {
	return &CheckResult{
		Name:    name,
		State:   StateFailed,
		Summary: summary,
		Details: map[string]interface{}{
			"error": err,
		},
	}
}

func WriteResults(out io.Writer, results CheckResults) {
	for _, result := range results {
		switch result.State {
		case StatePassed:
			io.WriteString(out, "PASS ")
		case StateWarning:
			io.WriteString(out, "WARN ")
		case StateFailed:
			io.WriteString(out, "FAIL ")
		case StateInfo:
			io.WriteString(out, "INFO ")
		case StateSkipped:
			io.WriteString(out, "SKIP ")
		}

		fmt.Fprintln(out, result.Summary)
	}
}
