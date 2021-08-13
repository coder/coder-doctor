package api

// ResultWriter writes the given result to a configured output.
type ResultWriter interface {
	WriteResult(result *CheckResult) error
}

// DiscardWriter is a writer that discards all results.
type DiscardWriter struct {
}

func (*DiscardWriter) WriteResult(result *CheckResult) error {
	return nil
}
