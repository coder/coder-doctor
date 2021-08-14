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

var _ = ResultWriter(&CaptureWriter{})

// CaptureWriter is a writer that stores all results in an
// internal buffer.
type CaptureWriter struct {
	results []*CheckResult
}

func (w *CaptureWriter) WriteResult(result *CheckResult) error {
	w.results = append(w.results, result)
	return nil
}

func (w *CaptureWriter) Clear() {
	w.results = nil
}

func (w *CaptureWriter) Get() []*CheckResult {
	return w.results
}

func (w *CaptureWriter) Empty() bool {
	return w.Len() == 0
}

func (w *CaptureWriter) Len() int {
	return len(w.results)
}
