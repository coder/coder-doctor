package kube

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/cdr/coder-doctor/internal/api"
)

func (k *KubernetesChecker) CheckResources(_ context.Context) []*api.CheckResult {
	const checkName = "kubernetes-resources"
	results := make([]*api.CheckResult, 0)
	dc := k.client.Discovery()
	lists, err := dc.ServerPreferredResources()
	if err != nil {
		results = append(results, api.SkippedResult(checkName, "unable to fetch api resources from server: "+err.Error()))
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

	versionReqs := findClosestVersionRequirements(k.coderVersion)
	for versionReq := range versionReqs {
		if !resourcesAvailable[*versionReq] {
			msg := fmt.Sprintf("missing required resource:%q group:%q version:%q", versionReq.Resource, versionReq.Group, versionReq.Version)
			errResult := api.ErrorResult(checkName, msg, xerrors.New(msg))
			results = append(results, errResult)
			continue
		}
		msg := fmt.Sprintf("found required resource:%q group:%q version:%q", versionReq.Resource, versionReq.Group, versionReq.Version)
		results = append(results, api.PassResult(checkName, msg))
	}

	return results
}
