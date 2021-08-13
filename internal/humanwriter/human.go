package humanwriter

import (
	"fmt"
	"io"

	"github.com/cdr/coder-doctor/internal/api"
)

type OutputMode int

const (
	OutputModeText OutputMode = iota
	OutputModeEmoji
)

func (m OutputMode) String() string {
	switch m {
	case OutputModeEmoji:
		return "OutputModeEmoji"
	case OutputModeText:
		return "OutputModeText"
	}

	panic(fmt.Sprintf("unknown OutputMode: %d", m))
}

type HumanResultWriter struct {
	out    io.Writer
	mode   OutputMode
	colors bool
}

type Option func(w *HumanResultWriter)

func New(out io.Writer, opts ...Option) *HumanResultWriter {
	w := &HumanResultWriter{
		out:    out,
		mode:   OutputModeText,
		colors: false,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func WithColors(colors bool) Option {
	return func(w *HumanResultWriter) {
		w.colors = colors
	}
}

func WithMode(mode OutputMode) Option {
	return func(w *HumanResultWriter) {
		w.mode = mode
	}
}

func (w *HumanResultWriter) WriteResult(result *api.CheckResult) error {
	var prefix string
	var err error

	switch w.mode {
	case OutputModeEmoji:
		prefix, err = result.State.Emoji()
	case OutputModeText:
		prefix, err = result.State.Text()
	}

	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w.out, prefix, result.Summary)
	return err
}
