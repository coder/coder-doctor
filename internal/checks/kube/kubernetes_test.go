package kube

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	"cdr.dev/slog/sloggers/slogtest/assert"

	"github.com/Masterminds/semver/v3"
)

func TestKubernetesOptions(t *testing.T) {
	t.Parallel()

	clientset := fake.NewSimpleClientset()
	checker := NewKubernetesChecker(clientset)
	assert.Success(t, "validation successful", checker.Validate())

	checker = NewKubernetesChecker(clientset, WithCoderVersion(semver.MustParse("1.19.0")))
	assert.ErrorContains(t, "KubernetesChecker with unknown version should fail to validate", checker.Validate(), "unhandled coder version")
	// var buf bytes.Buffer
	// log := slog.Make(sloghuman.Sink(&buf)).Leveled(slog.LevelDebug)
	// checker = NewKubernetesChecker(clientset,
	// 	WithCoderVersion(semver.MustParse("1.0")),
	// 	WithLogger(log))
	// assert.True(t, "log has output", buf.Len() > 0)
}
