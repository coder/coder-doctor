package filterwriter

import (
	"github.com/cdr/coder-doctor/internal/api"
)

var _ = api.ResultWriter(&FilterWriter{})

type FilterWriter struct {
	writer api.ResultWriter
	filter int
}

type Option func(w *FilterWriter)

func Must(writer api.ResultWriter, opts ...Option) *FilterWriter {
	filter, err := New(writer, opts...)
	if err != nil {
		panic(err.Error())
	}

	return filter
}

func New(writer api.ResultWriter, opts ...Option) (*FilterWriter, error) {
	w := &FilterWriter{
		writer: writer,
	}

	// The default is to allow Pass, Warn, and Fail results only
	WithAcceptState(api.StatePassed)(w)
	WithAcceptState(api.StateWarning)(w)
	WithAcceptState(api.StateFailed)(w)

	for _, opt := range opts {
		opt(w)
	}

	return w, nil
}

func WithAcceptState(level api.CheckState) Option {
	return func(w *FilterWriter) {
		w.filter |= 1 << level
	}
}

func WithFilterState(level api.CheckState) Option {
	return func(w *FilterWriter) {
		w.filter &= ^(1 << level)
	}
}

func (w *FilterWriter) WriteResult(result *api.CheckResult) error {
	if w.filter&(1<<result.State) != 0 {
		return w.writer.WriteResult(result)
	}

	return nil
}
