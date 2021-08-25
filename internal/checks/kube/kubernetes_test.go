package kube

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	"cdr.dev/slog/sloggers/slogtest/assert"

	"github.com/Masterminds/semver/v3"
)

func TestKubernetesOptions(t *testing.T) {
	t.Parallel()

	t.Run("successful validation", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		checker := NewKubernetesChecker(clientset)
		assert.Success(t, "validation successful", checker.Validate())
	})

	t.Run("validation failed: unhandled coder version", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("expected a panic")
				t.FailNow()
			}
			assert.ErrorContains(t, "KubernetesChecker with unknown version should fail to validate", r.(error), "unhandled coder version")
		}()

		// This should panic
		_ = NewKubernetesChecker(fake.NewSimpleClientset(), WithCoderVersion(semver.MustParse("1.19.0")))
	})

	// var buf bytes.Buffer
	// log := slog.Make(sloghuman.Sink(&buf)).Leveled(slog.LevelDebug)
	// checker = NewKubernetesChecker(clientset,
	// 	WithCoderVersion(semver.MustParse("1.0")),
	// 	WithLogger(log))
	// assert.True(t, "log has output", buf.Len() > 0)
}
