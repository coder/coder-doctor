package checks

import (
	"context"

	"github.com/cdr/doctor/internal/api"
	"github.com/cdr/doctor/internal/checks/kubernetes"
)

var kubernetesChecks = []api.Check{
	kubernetes.CheckVersion,
}

func RunKubernetes(ctx context.Context, opts api.CheckOptions) api.CheckResults {
	return kubernetes.CheckVersion(ctx, opts)
}
