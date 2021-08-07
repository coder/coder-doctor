package kubernetes

import (
	"context"
	"fmt"

	"github.com/cdr/coder-doctor/internal/api"
)

func CheckVersion(ctx context.Context, opts api.CheckOptions) api.CheckResults {
	// coderVersion := opts.CoderVersion
	client := opts.Kubernetes

	versionInfo, err := client.Discovery().ServerVersion()
	if err != nil {
		return api.CheckResults{
			api.CheckResult{
				Name:    "kubernetes-version",
				State:   api.StateFailed,
				Summary: "failed to get Kubernetes version from server",
				Details: map[string]interface{}{
					"error": err,
				},
			},
		}
	}

	return api.CheckResults{
		api.CheckResult{
			Name:    "kubernetes-version",
			State:   api.StateInfo,
			Summary: fmt.Sprintf("kubernetes version: %s", versionInfo),
		},
	}
}
