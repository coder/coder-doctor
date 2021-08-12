package kube

import (
	"context"
	"io"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"
	"github.com/Masterminds/semver/v3"
	"k8s.io/client-go/kubernetes"

	"github.com/cdr/coder-doctor/internal/api"
)

var _ = api.Checker(&KubernetesChecker{})

type KubernetesChecker struct {
	client       kubernetes.Interface
	coderVersion *semver.Version
	log          slog.Logger
}

type KubernetesCheckOption func(k *KubernetesChecker)

func NewKubernetesChecker(opts ...KubernetesCheckOption) *KubernetesChecker {
	checker := &KubernetesChecker{
		log: slog.Make(sloghuman.Sink(io.Discard)),
	}

	for _, opt := range opts {
		opt(checker)
	}

	return checker
}

func (*KubernetesChecker) Validate() error {
	return nil
}

func (k *KubernetesChecker) Run(ctx context.Context) api.CheckResults {
	return api.CheckResults{
		k.CheckVersion(ctx),
	}
}

func WithClient(client kubernetes.Interface) KubernetesCheckOption {
	return func(k *KubernetesChecker) {
		k.client = client
	}
}

func WithCoderVersion(version *semver.Version) KubernetesCheckOption {
	return func(k *KubernetesChecker) {
		k.coderVersion = version
	}
}

func WithLogger(log slog.Logger) KubernetesCheckOption {
	return func(k *KubernetesChecker) {
		k.log = log
	}
}
