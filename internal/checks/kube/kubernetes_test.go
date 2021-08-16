package kube

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	"cdr.dev/slog/sloggers/slogtest/assert"
)

func TestKubernetesOptions(t *testing.T) {
	t.Parallel()

	clientset := fake.NewSimpleClientset()
	checker := NewKubernetesChecker(clientset)
	assert.Success(t, "validation successful", checker.Validate())

	// var buf bytes.Buffer
	// log := slog.Make(sloghuman.Sink(&buf)).Leveled(slog.LevelDebug)
	// checker = NewKubernetesChecker(clientset,
	// 	WithCoderVersion(semver.MustParse("1.0")),
	// 	WithLogger(log))
	// assert.True(t, "log has output", buf.Len() > 0)
}
