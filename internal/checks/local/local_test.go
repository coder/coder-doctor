package local

import (
	"context"
	"strings"
	"testing"
)

type execResult struct {
	Output []byte
	Err    error
}

func newFakeExecer(t *testing.T) *fakeExecer {
	m := make(map[string]execResult)
	return &fakeExecer{
		M: m,
		T: t,
	}
}

type fakeExecer struct {
	M map[string]execResult
	T *testing.T
}

func (f *fakeExecer) Handle(cmd string, output []byte, err error) {
	f.M[cmd] = execResult{
		Output: output,
		Err:    err,
	}
}

func (f *fakeExecer) ExecContext(_ context.Context, name string, args ...string) ([]byte, error) {
	var sb strings.Builder
	_, _ = sb.WriteString(name)
	for _, arg := range args {
		_, _ = sb.WriteString(" ")
		_, _ = sb.WriteString(arg)
	}

	fullCmd := sb.String()
	res, ok := f.M[fullCmd]
	if !ok {
		f.T.Logf("unhandled ExecContext: %s", fullCmd)
		f.T.FailNow()
		return nil, nil // should never happen
	}

	return res.Output, res.Err
}

type lookPathResult struct {
	S   string
	Err error
}

type fakeLookPather struct {
	M map[string]lookPathResult
	T *testing.T
}

func (f *fakeLookPather) LookPath(name string) (string, error) {
	res, ok := f.M[name]
	if !ok {
		f.T.Logf("unhandled LookPath: %s", name)
		f.T.FailNow()
	}

	return res.S, res.Err
}

func (f *fakeLookPather) Handle(name string, path string, err error) {
	f.M[name] = lookPathResult{
		S:   path,
		Err: err,
	}
}

func newFakeLookPather(t *testing.T) *fakeLookPather {
	m := make(map[string]lookPathResult)
	return &fakeLookPather{
		M: m,
		T: t,
	}
}
