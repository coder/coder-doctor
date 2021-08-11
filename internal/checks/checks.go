package checks

import (
	"context"

	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/checks/kubernetes"
)

func RunKubernetes(ctx context.Context, opts api.CheckOptions) api.CheckResults {
	return api.CheckResults{
		kubernetes.CheckVersion(ctx, opts),
	}
}
