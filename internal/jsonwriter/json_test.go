package jsonwriter_test

import (
	"strings"
	"testing"

	"cdr.dev/coder-doctor/internal/api"
	"cdr.dev/coder-doctor/internal/jsonwriter"
	"cdr.dev/slog/sloggers/slogtest/assert"
)

func TestJSONWriter(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	w := jsonwriter.New(&buf)
	w.WriteResult(&api.CheckResult{
		State:   api.StateFailed,
		Name:    "checks",
		Summary: "test check",
		Details: map[string]interface{}{
			"number": 123,
			"string": "hello world",
		},
	})
	expected := `{"name":"checks","state":2,"summary":"test check","details":{"number":123,"string":"hello world"}}` + "\n"
	assert.Equal(t, "check with details", expected, buf.String())
	buf.Reset()

	w.WriteResult(&api.CheckResult{
		State:   api.StatePassed,
		Name:    "checks",
		Summary: "test passing check",
	})
	expected = `{"name":"checks","state":0,"summary":"test passing check"}` + "\n"
	assert.Equal(t, "check without details", expected, buf.String())
}
