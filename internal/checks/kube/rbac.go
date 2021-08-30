package kube

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"github.com/Masterminds/semver/v3"

	"github.com/cdr/coder-doctor/internal/api"

	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationclientv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	rbacutil "k8s.io/kubectl/pkg/util/rbac"
	"k8s.io/kubectl/pkg/util/slice"
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

func (k *KubernetesChecker) CheckRBACSSRR(ctx context.Context) []*api.CheckResult {
	const checkName = "kubernetes-rbac-ssrr"
	authClient := k.client.AuthorizationV1()
	results := make([]*api.CheckResult, 0)

	sar := &authorizationv1.SelfSubjectRulesReview{
		Spec: authorizationv1.SelfSubjectRulesReviewSpec{
			Namespace: k.namespace,
		},
	}

	response, err := authClient.SelfSubjectRulesReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return []*api.CheckResult{api.SkippedResult(checkName, "unable to create SelfSubjectRulesReview request: "+err.Error())}
	}

	if response.Status.Incomplete {
		results = append(results, api.WarnResult(checkName, "SelfSubjectRulesReview response incomplete - results may not be correct"))
	}

	// convert the list of rules from the server to PolicyRules and dedupe/compact
	breakdownRules := []rbacv1.PolicyRule{}
	for _, rule := range convertToPolicyRule(response.Status) {
		breakdownRules = append(breakdownRules, rbacutil.BreakdownRule(rule)...)
	}

	compactRules, _ := rbacutil.CompactRules(breakdownRules) //nolint:errcheck - err is always nil
	sort.Stable(rbacutil.SortableRuleSlice(compactRules))

	// TODO: optimise this
	for req, reqVerbs := range k.reqs.ResourceRequirements {
		if err := satisfies(req, reqVerbs, compactRules); err != nil {
			summary := fmt.Sprintf("missing permissions on resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(checkName, summary, err))
			continue
		}
		summary := fmt.Sprintf("%s: can %s", req.Resource, strings.Join(reqVerbs, ", "))
		results = append(results, api.PassResult(checkName, summary))
	}

	return results
}

func satisfies(req *ResourceRequirement, verbs ResourceVerbs, rules []rbacv1.PolicyRule) error {
	for _, rule := range rules {
		if apiGroupsMatch(req.Group, rule.APIGroups) &&
			apiResourceMatch(req.Resource, rule.Resources) &&
			verbsMatch(verbs, rule.Verbs) {
			return nil
		}
	}
	return xerrors.Errorf("not satisfied: group:%q resource:%q version:%q verbs:%+v", req.Group, req.Resource, req.Version, verbs)
}

// The below adapted from k8s.io/pkg/apis/rbac/v1/evaluation_helpers.go
func verbsMatch(want, have ResourceVerbs) bool {
	if slice.ContainsString(have, "*", nil) {
		return true
	}

	for _, v := range want {
		if !slice.ContainsString(have, v, nil) {
			return false
		}
	}
	return true
}

func apiGroupsMatch(want string, have []string) bool {
	for _, g := range have {
		if g == "*" {
			return true
		}

		if g == want {
			return true
		}
	}
	return false
}

func apiResourceMatch(want string, have []string) bool {
	for _, r := range have {
		if r == "*" {
			return true
		}

		if r == want {
			return true
		}
	}
	return false
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

// lifted from kubectl/pkg/cmd/auth/cani.go
func convertToPolicyRule(status authorizationv1.SubjectRulesReviewStatus) []rbacv1.PolicyRule {
	ret := []rbacv1.PolicyRule{}
	for _, resource := range status.ResourceRules {
		ret = append(ret, rbacv1.PolicyRule{
			Verbs:         resource.Verbs,
			APIGroups:     resource.APIGroups,
			Resources:     resource.Resources,
			ResourceNames: resource.ResourceNames,
		})
	}

	for _, nonResource := range status.NonResourceRules {
		ret = append(ret, rbacv1.PolicyRule{
			Verbs:           nonResource.Verbs,
			NonResourceURLs: nonResource.NonResourceURLs,
		})
	}

	return ret
}
