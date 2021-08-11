package kubernetes

import (
	"fmt"
	"io"

	"github.com/cdr/coder-doctor/internal/api"
)

func PrintResults(out io.Writer, results api.CheckResults) {
	for _, result := range results {
		if result.State == api.StatePassed {
			fmt.Fprintf(out, "PASS ")
		}
		fmt.Fprintln(out, result.Summary)
	}
}
