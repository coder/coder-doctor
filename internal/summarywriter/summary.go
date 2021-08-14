package summarywriter

import "github.com/cdr/coder-doctor/internal/api"

type SummaryResult struct {
	Passed  int `json:"passed"`
	Warning int `json:"warning"`
	Failed  int `json:"failed"`
	Info    int `json:"info"`
	Skipped int `json:"skipped"`
	Total   int `json:"total"`
}

var _ = api.ResultWriter(&SummaryWriter{})

type SummaryWriter struct {
	summary SummaryResult
	writer  api.ResultWriter
}

func New(writer api.ResultWriter) *SummaryWriter {
	return &SummaryWriter{
		writer: writer,
	}
}

func (w *SummaryWriter) WriteResult(result *api.CheckResult) error {
	w.summary.Total++

	switch result.State {
	case api.StatePassed:
		w.summary.Passed++
	case api.StateWarning:
		w.summary.Warning++
	case api.StateFailed:
		w.summary.Failed++
	case api.StateInfo:
		w.summary.Info++
	case api.StateSkipped:
		w.summary.Skipped++
	}

	return w.writer.WriteResult(result)
}

func (w *SummaryWriter) Reset() {
	w.summary = SummaryResult{}
}

func (w *SummaryWriter) Summary() SummaryResult {
	return w.summary
}
