package kube

import (
	"context"

	"github.com/cdr/coder-doctor/internal/api"
)

func (k *KubernetesChecker) CheckResources(ctx context.Context) []*api.CheckResult {
	results := make([]*api.CheckResult, 0)
	return results
}
