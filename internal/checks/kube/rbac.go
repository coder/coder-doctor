package kube

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/cdr/coder-doctor/internal/api"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationv1client "k8s.io/client-go/kubernetes/typed/authorization/v1" //nolint:importas
)

type RBACRequirement struct {
	Resource string
	Verbs    []string
}

var VerbsCreateDeleteList = []string{"create", "delete", "list"}

func NewRBACRequirement(resource string, verbs ...string) *RBACRequirement {
	return &RBACRequirement{
		Resource: resource,
		Verbs:    verbs,
	}
}

var rbacRequirements = []*RBACRequirement{
	NewRBACRequirement("deployments", VerbsCreateDeleteList...),
	NewRBACRequirement("serviceaccounts", VerbsCreateDeleteList...),
	NewRBACRequirement("replicasets", VerbsCreateDeleteList...),
	NewRBACRequirement("pods", VerbsCreateDeleteList...),
	NewRBACRequirement("roles", VerbsCreateDeleteList...),
	NewRBACRequirement("rolebindings", VerbsCreateDeleteList...),
	NewRBACRequirement("ingresses", VerbsCreateDeleteList...),
	NewRBACRequirement("secrets", VerbsCreateDeleteList...),
	NewRBACRequirement("services", VerbsCreateDeleteList...),
	NewRBACRequirement("statefulsets", VerbsCreateDeleteList...),
}

func (k *KubernetesChecker) CheckRBAC(ctx context.Context) []*api.CheckResult {
	const checkName = "kubernetes-rbac"
	authClient := k.client.AuthorizationV1()
	results := make([]*api.CheckResult, 0, len(rbacRequirements))

	for _, req := range rbacRequirements {
		resName := fmt.Sprintf("%s-%s", checkName, req.Resource)
		if err := k.checkOneRBAC(ctx, authClient, req); err != nil {
			summary := fmt.Sprintf("missing permissions on resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(resName, summary, err))
			continue
		}

		summary := fmt.Sprintf("%s: can %s", req.Resource, strings.Join(req.Verbs, ", "))
		results = append(results, api.PassResult(resName, summary))
	}

	return results
}

func (k *KubernetesChecker) checkOneRBAC(ctx context.Context, authClient authorizationv1client.AuthorizationV1Interface, req *RBACRequirement) error {
	have := make([]string, 0, len(req.Verbs))
	for _, verb := range req.Verbs {
		sar := &authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Namespace: k.namespace,
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

	if len(have) != len(req.Verbs) {
		return xerrors.Errorf(fmt.Sprintf("need: %+v have: %+v", req.Verbs, have))
	}

	return nil
}
