package local

import (
	"context"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

func Test_CheckLocalHelmVersion(t *testing.T) {
	t.Parallel()

	type params struct {
		W    *api.CaptureWriter
		EX   *fakeExecer
		LP   *fakeLookPather
		Opts []Option
		Ctx  context.Context
	}

	run := func(t *testing.T, name string, fn func(t *testing.T, p *params)) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			cw := &api.CaptureWriter{}
			ex := newFakeExecer(t)
			lp := newFakeLookPather(t)
			opts := []Option{
				WithWriter(cw),
				WithExecF(ex.ExecContext),
				WithLookPathF(lp.LookPath),
				WithTarget(api.CheckTargetKubernetes), // default
			}
			p := &params{
				W:    cw,
				EX:   ex,
				LP:   lp,
				Opts: opts,
				Ctx:  ctx,
			}
			fn(t, p)
		})
	}

	run(t, "helm: when not running against kubernetes", func(t *testing.T, p *params) {
		p.Opts = append(p.Opts, WithTarget(api.CheckTargetUndefined))
		lc := NewChecker(p.Opts...)
		err := lc.Run(p.Ctx)
		assert.Success(t, "run local checker", err)
		assert.False(t, "results should not be empty", p.W.Empty())
		for _, res := range p.W.Get() {
			if res.Name == LocalHelmVersionCheck {
				assert.Equal(t, "should skip helm check if not running against kubernetes", api.StateSkipped, res.State)
			}
		}
	})

	run(t, "helm: with version 3.6", func(t *testing.T, p *params) {
		p.LP.Handle("helm", "/usr/local/bin/helm", nil)
		p.EX.Handle("/usr/local/bin/helm version --short", []byte("v3.6.0+g7f2df64"), nil)
		lc := NewChecker(p.Opts...)
		err := lc.Run(p.Ctx)
		assert.Success(t, "run local checker", err)
		assert.False(t, "results should not be empty", p.W.Empty())
		for _, res := range p.W.Get() {
			if res.Name == LocalHelmVersionCheck {
				assert.Equal(t, "should pass", api.StatePassed, res.State)
			}
		}
	})

	run(t, "helm: with version 2", func(t *testing.T, p *params) {
		p.LP.Handle("helm", "/usr/local/bin/helm", nil)
		p.EX.Handle("/usr/local/bin/helm version --short", []byte("v2.0.0"), nil)
		lc := NewChecker(p.Opts...)
		err := lc.Run(p.Ctx)
		assert.Success(t, "run local checker", err)
		assert.False(t, "results should not be empty", p.W.Empty())
		for _, res := range p.W.Get() {
			if res.Name == LocalHelmVersionCheck {
				assert.Equal(t, "should fail", api.StateFailed, res.State)
			}
		}
	})

	run(t, "helm: not in path", func(t *testing.T, p *params) {
		p.LP.Handle("helm", "", os.ErrNotExist)
		lc := NewChecker(p.Opts...)
		err := lc.Run(p.Ctx)
		assert.Success(t, "run local checker", err)
		assert.False(t, "results should not be empty", p.W.Empty())
		for _, res := range p.W.Get() {
			if res.Name == LocalHelmVersionCheck {
				assert.Equal(t, "should fail", api.StateFailed, res.State)
			}
		}
	})

	run(t, "helm: cannot be executed", func(t *testing.T, p *params) {
		p.LP.Handle("helm", "/usr/local/bin/helm", nil)
		p.EX.Handle("/usr/local/bin/helm version --short", []byte(""), os.ErrPermission)
		lc := NewChecker(p.Opts...)
		err := lc.Run(p.Ctx)
		assert.Success(t, "run local checker", err)
		assert.False(t, "results should not be empty", p.W.Empty())
		for _, res := range p.W.Get() {
			if res.Name == LocalHelmVersionCheck {
				assert.Equal(t, "should fail", api.StateFailed, res.State)
			}
		}
	})

	run(t, "helm: returns garbage version", func(t *testing.T, p *params) {
		p.LP.Handle("helm", "/usr/local/bin/helm", nil)
		p.EX.Handle("/usr/local/bin/helm version --short", []byte(""), nil)
		lc := NewChecker(p.Opts...)
		err := lc.Run(p.Ctx)
		assert.Success(t, "run local checker", err)
		assert.False(t, "results should not be empty", p.W.Empty())
		for _, res := range p.W.Get() {
			if res.Name == LocalHelmVersionCheck {
				assert.Equal(t, "should fail", api.StateFailed, res.State)
			}
		}
	})

	run(t, "helm: someone did not call validate", func(t *testing.T, p *params) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("this code should have panicked")
				t.FailNow()
			}
		}()
		p.Opts = append(p.Opts, WithCoderVersion(semver.MustParse("v1.19")))
		p.LP.Handle("helm", "/usr/local/bin/helm", nil)
		p.EX.Handle("/usr/local/bin/helm version --short", []byte("v3.6.0+g7f2df64"), nil)
		lc := NewChecker(p.Opts...)
		_ = lc.Run(p.Ctx)
	})
}
