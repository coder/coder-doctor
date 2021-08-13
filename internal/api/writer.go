package api

var _ = ResultWriter(&DiscardWriter{})

// ResultWriter writes the given result to a configured output.
type ResultWriter interface {
	WriteResult(*CheckResult) error
}

// DiscardWriter is a writer that discards all results.
type DiscardWriter struct {
}

func (*DiscardWriter) WriteResult(_ *CheckResult) error {
	return nil
}
