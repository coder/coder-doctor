package jsonwriter

import (
	"encoding/json"
	"io"

	"cdr.dev/coder-doctor/internal/api"
)

var _ = api.ResultWriter(&JSONWriter{})

// JSONWriter is a writer that writes results to a stream in
// JSON Lines format.
type JSONWriter struct {
	writer  io.Writer
	encoder *json.Encoder
}

func New(writer io.Writer) *JSONWriter {
	return &JSONWriter{
		writer:  writer,
		encoder: json.NewEncoder(writer),
	}
}

func (w *JSONWriter) WriteResult(result *api.CheckResult) error {
	return w.encoder.Encode(result)
}
