package kube

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/cdr/coder-doctor/internal/api"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationclientv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type RBACRequirement struct {
	APIGroup string
	Resource string
	Verbs    []string
}

var VerbsCreateDeleteList = []string{"create", "delete", "list"}

func NewRBACRequirement(apiGroup, resource string, verbs ...string) *RBACRequirement {
	return &RBACRequirement{
		APIGroup: apiGroup,
		Resource: resource,
		Verbs:    verbs,
	}
}

var rbacRequirements = []*RBACRequirement{
	NewRBACRequirement("", "pods", VerbsCreateDeleteList...),
	NewRBACRequirement("", "roles", VerbsCreateDeleteList...),
	NewRBACRequirement("", "rolebindings", VerbsCreateDeleteList...),
	NewRBACRequirement("", "secrets", VerbsCreateDeleteList...),
	NewRBACRequirement("", "serviceaccounts", VerbsCreateDeleteList...),
	NewRBACRequirement("", "services", VerbsCreateDeleteList...),
	NewRBACRequirement("apps", "deployments", VerbsCreateDeleteList...),
	NewRBACRequirement("apps", "replicasets", VerbsCreateDeleteList...),
	NewRBACRequirement("apps", "statefulsets", VerbsCreateDeleteList...),
	NewRBACRequirement("extensions", "ingresses", VerbsCreateDeleteList...),
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

func (k *KubernetesChecker) checkOneRBAC(ctx context.Context, authClient authorizationclientv1.AuthorizationV1Interface, req *RBACRequirement) error {
	have := make([]string, 0, len(req.Verbs))
	for _, verb := range req.Verbs {
		sar := &authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Namespace: k.namespace,
					Group:     req.APIGroup,
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
