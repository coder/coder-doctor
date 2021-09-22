package local

import (
	"context"
	"io"
	"os/exec"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/xerrors"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"

	"cdr.dev/coder-doctor/internal/api"
)

var _ api.Checker = &Checker{}

type ExecF func(ctx context.Context, name string, args ...string) ([]byte, error)
type LookPathF func(string) (string, error)

// local.Checker checks the local environment.
type Checker struct {
	writer       api.ResultWriter
	coderVersion *semver.Version
	log          slog.Logger
	target       api.CheckTarget
	execF        ExecF
	lookPathF    LookPathF
}

type Option func(*Checker)

func NewChecker(opts ...Option) *Checker {
	checker := &Checker{
		writer:       &api.DiscardWriter{},
		coderVersion: semver.MustParse("100.0.0"),
		log:          slog.Make(sloghuman.Sink(io.Discard)),
		execF:        defaultExecCommand,
		lookPathF:    exec.LookPath,
	}

	for _, opt := range opts {
		opt(checker)
	}

	if err := checker.Validate(); err != nil {
		panic(xerrors.Errorf("error validating local checker: %w", err))
	}

	return checker
}

func WithTarget(t api.CheckTarget) Option {
	return func(l *Checker) {
		l.target = t
	}
}

func WithWriter(writer api.ResultWriter) Option {
	return func(l *Checker) {
		l.writer = writer
	}
}

func WithCoderVersion(version *semver.Version) Option {
	return func(l *Checker) {
		l.coderVersion = version
	}
}

func WithLogger(log slog.Logger) Option {
	return func(l *Checker) {
		l.log = log
	}
}

func WithExecF(f ExecF) Option {
	return func(l *Checker) {
		l.execF = f
	}
}

func WithLookPathF(f LookPathF) Option {
	return func(l *Checker) {
		l.lookPathF = f
	}
}

func (l *Checker) Validate() error {
	// Ensure we know the Helm version requirement for our Coder version.
	if findNearestHelmVersion(l.coderVersion) == nil {
		return xerrors.Errorf("unhandled coder version %s: compatible helm version not specified", l.coderVersion.String())
	}
	return nil
}

func (l *Checker) Run(ctx context.Context) error {
	if err := l.writer.WriteResult(l.CheckLocalHelmVersion(ctx)); err != nil {
		return xerrors.Errorf("check local helm version: %w", err)
	}
	return nil
}

func defaultExecCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("exec %q %+q: %w", name, args, err)
	}

	return out, nil
}
