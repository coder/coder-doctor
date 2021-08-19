package kube

import (
	"context"
	"io"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/xerrors"
	"k8s.io/client-go/kubernetes"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"

	"github.com/cdr/coder-doctor/internal/api"
)

var _ = api.Checker(&KubernetesChecker{})

type KubernetesChecker struct {
	namespace    string
	client       kubernetes.Interface
	writer       api.ResultWriter
	coderVersion *semver.Version
	log          slog.Logger
}

type Option func(k *KubernetesChecker)

func NewKubernetesChecker(client kubernetes.Interface, opts ...Option) *KubernetesChecker {
	checker := &KubernetesChecker{
		namespace: "default",
		client:    client,
		log:       slog.Make(sloghuman.Sink(io.Discard)),
		writer:    &api.DiscardWriter{},
		// Select the newest version by default
		coderVersion: semver.MustParse("100.0.0"),
	}

	for _, opt := range opts {
		opt(checker)
	}

	return checker
}

func WithWriter(writer api.ResultWriter) Option {
	return func(k *KubernetesChecker) {
		k.writer = writer
	}
}

func WithCoderVersion(version *semver.Version) Option {
	return func(k *KubernetesChecker) {
		k.coderVersion = version
	}
}

func WithLogger(log slog.Logger) Option {
	return func(k *KubernetesChecker) {
		k.log = log
	}
}

func WithNamespace(ns string) Option {
	return func(k *KubernetesChecker) {
		k.namespace = ns
	}
}

func (*KubernetesChecker) Validate() error {
	return nil
}

func (k *KubernetesChecker) Run(ctx context.Context) error {
	err := k.writer.WriteResult(k.CheckVersion(ctx))
	if err != nil {
		return xerrors.Errorf("check version: %w", err)
	}

	for _, res := range k.CheckRBAC(ctx) {
		if err := k.writer.WriteResult(res); err != nil {
			return xerrors.Errorf("check RBAC: %w", err)
		}
	}
	return nil
}
