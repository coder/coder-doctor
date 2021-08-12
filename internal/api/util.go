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

func WriteResults(out io.Writer, results CheckResults) error {
	var err error
	for _, result := range results {
		switch result.State {
		case StatePassed:
			_, err = io.WriteString(out, "PASS ")
		case StateWarning:
			_, err = io.WriteString(out, "WARN ")
		case StateFailed:
			_, err = io.WriteString(out, "FAIL ")
		case StateInfo:
			_, err = io.WriteString(out, "INFO ")
		case StateSkipped:
			_, err = io.WriteString(out, "SKIP ")
		}

		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(out, result.Summary)
		if err != nil {
			return err
		}
	}

	return nil
}
