package kube

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/Masterminds/semver/v3"

	"github.com/cdr/coder-doctor/internal/api"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationclientv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

func (k *KubernetesChecker) CheckRBAC(ctx context.Context) []*api.CheckResult {
	const checkName = "kubernetes-rbac"
	authClient := k.client.AuthorizationV1()
	results := make([]*api.CheckResult, 0)

	for req, reqVerbs := range k.reqs.ResourceRequirements {
		resName := fmt.Sprintf("%s-%s", checkName, req.Resource)
		if err := k.checkOneRBAC(ctx, authClient, req, reqVerbs); err != nil {
			summary := fmt.Sprintf("missing permissions on resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(resName, summary, err))
			continue
		}

		summary := fmt.Sprintf("%s: can %s", req.Resource, strings.Join(reqVerbs, ", "))
		results = append(results, api.PassResult(resName, summary))
	}

	// TODO: delete this when the enterprise-helm role no longer requests resources on things
	// that don't exist.
	for req, reqVerbs := range k.reqs.RoleOnlyResourceRequirements {
		resName := fmt.Sprintf("%s-%s", checkName, req.Resource)
		if err := k.checkOneRBAC(ctx, authClient, req, reqVerbs); err != nil {
			summary := fmt.Sprintf("missing permissions on resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(resName, summary, err))
			continue
		}

		summary := fmt.Sprintf("%s: can %s", req.Resource, strings.Join(reqVerbs, ", "))
		results = append(results, api.PassResult(resName, summary))
	}

	return results
}

func (k *KubernetesChecker) checkOneRBAC(ctx context.Context, authClient authorizationclientv1.AuthorizationV1Interface, req *ResourceRequirement, reqVerbs ResourceVerbs) error {
	have := make([]string, 0, len(reqVerbs))
	for _, verb := range reqVerbs {
		sar := &authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Namespace: k.namespace,
					Group:     req.Group,
					Resource:  req.Resource,
					Verb:      verb,
				},
			},
		}

		response, err := authClient.SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})

		if err != nil {
			// should not fail - short-circuit
			return xerrors.Errorf("failed to create SelfSubjectAccessReview request: %w", err)
		}

		if response.Status.Allowed {
			have = append(have, verb)
			continue
		}
	}

	if len(have) != len(reqVerbs) {
		return xerrors.Errorf(fmt.Sprintf("need: %+v have: %+v", reqVerbs, have))
	}

	return nil
}

func findClosestVersionRequirements(v *semver.Version) *VersionedResourceRequirements {
	for _, vreqs := range allRequirements {
		if vreqs.VersionConstraints.Check(v) {
			return &vreqs
		}
	}
	return nil
}
