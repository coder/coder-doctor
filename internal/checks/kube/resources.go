package kube

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"cdr.dev/coder-doctor/internal/api"
)

func (k *KubernetesChecker) CheckResources(_ context.Context) []*api.CheckResult {
	const checkName = "kubernetes-resources"
	results := make([]*api.CheckResult, 0)
	dc := k.client.Discovery()
	lists, err := dc.ServerPreferredResources()
	if err != nil {
		results = append(results, api.SkippedResult(checkName, "unable to fetch api resources from server", err))
		return results
	}

	resourcesAvailable := make(map[ResourceRequirement]bool)
	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}

		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range list.APIResources {
			if len(resource.Verbs) == 0 {
				continue
			}

			r := ResourceRequirement{
				Group:    gv.Group,
				Version:  gv.String(),
				Resource: resource.Name,
			}
			resourcesAvailable[r] = true
		}
	}

	for versionReq := range k.reqs.ResourceRequirements {
		result := &api.CheckResult{
			Summary: checkName,
			Details: map[string]interface{}{
				"resource":     versionReq.Resource,
				"group":        versionReq.Group,
				"groupVersion": versionReq.Version,
			},
		}

		if resourcesAvailable[*versionReq] {
			result.Summary = fmt.Sprintf("Cluster supports %s resource %s", versionReq.Version, versionReq.Resource)
			result.State = api.StatePassed
		} else {
			result.Summary = fmt.Sprintf("Cluster does not support %s resource %s", versionReq.Version, versionReq.Resource)
			result.State = api.StateFailed
		}
		results = append(results, result)
	}

	return results
}
