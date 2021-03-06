package kube

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"cdr.dev/slog"

	"github.com/Masterminds/semver/v3"

	"cdr.dev/coder-doctor/internal/api"

	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	rbacutil "k8s.io/kubectl/pkg/util/rbac"
	"k8s.io/kubectl/pkg/util/slice"
)

var errSelfSubjectRulesReviewNotSupported = xerrors.New("cluster does not support SelfSubjectRulesReview")

// CheckRBAC checks the cluster for the RBAC permissions required by Coder.
// It will attempt to first use a SelfSubjectRulesReview to determine the capabilities
// of the user. If this fails (notably on GKE), fall back to using SelfSubjectAccessRequests
// which is slower but is more likely to work.
func (k *KubernetesChecker) CheckRBAC(ctx context.Context) []*api.CheckResult {
	ssrrResults, err := k.checkRBACDefault(ctx)
	if err == nil {
		return ssrrResults
	}

	if xerrors.Is(err, errSelfSubjectRulesReviewNotSupported) {
		// In this case, we should fall back to using SelfSubjectAccessRequests.
		k.log.Warn(ctx, "unable to check via SelfSubjectRulesReview, falling back to SelfSubjectAccessRequests (slow)")
		return k.checkRBACFallback(ctx)
	}

	// something else went wrong
	return []*api.CheckResult{api.ErrorResult("kubernetes-rbac", "unable to check rbac", err)}
}

func (k *KubernetesChecker) checkRBACDefault(ctx context.Context) ([]*api.CheckResult, error) {
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
		return nil, xerrors.Errorf("create SelfSubjectRulesReview: %w", err)
	}

	if response.Status.Incomplete {
		return nil, errSelfSubjectRulesReviewNotSupported
	}

	// convert the list of rules from the server to PolicyRules and dedupe/compact
	breakdownRules := []rbacv1.PolicyRule{}
	for _, rule := range convertToPolicyRule(response.Status) {
		breakdownRules = append(breakdownRules, rbacutil.BreakdownRule(rule)...)
	}

	compactRules, err := rbacutil.CompactRules(breakdownRules)
	if err != nil {
		return nil, xerrors.Errorf("compact rules: %w", err)
	}

	sort.Stable(rbacutil.SortableRuleSlice(compactRules))
	for _, r := range compactRules {
		k.log.Debug(ctx, "Got SSRR PolicyRule", slog.F("rule", r))
	}

	// TODO: optimize this
	for req, reqVerbs := range k.reqs.ResourceRequirements {
		if err := satisfies(req, reqVerbs, compactRules); err != nil {
			summary := fmt.Sprintf("resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(checkName, summary, err))
			continue
		}
		resourceName := req.Resource
		if req.Group != "" {
			resourceName = req.Group + "/" + req.Resource
		}
		summary := fmt.Sprintf("resource %s: can %s", resourceName, strings.Join(reqVerbs, ", "))
		results = append(results, api.PassResult(checkName, summary))
	}

	// TODO: remove this when the helm chart is fixed
	for req, reqVerbs := range k.reqs.RoleOnlyResourceRequirements {
		if err := satisfies(req, reqVerbs, compactRules); err != nil {
			summary := fmt.Sprintf("resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(checkName, summary, err))
			continue
		}
		resourceName := req.Resource
		if req.Group != "" {
			resourceName = req.Group + "/" + req.Resource
		}
		summary := fmt.Sprintf("resource %s: can %s", resourceName, strings.Join(reqVerbs, ", "))
		results = append(results, api.PassResult(checkName, summary))
	}

	return results, nil
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

// checkRBACFallback uses a SelfSubjectAccessRequest to check the cluster for the required
// accesses. This requires a number of checks and is relatively slow.
func (k *KubernetesChecker) checkRBACFallback(ctx context.Context) []*api.CheckResult {
	const checkName = "kubernetes-rbac"
	authClient := k.client.AuthorizationV1()
	results := make([]*api.CheckResult, 0)

	for req, reqVerbs := range k.reqs.ResourceRequirements {
		if err := k.checkOneRBACSSAR(ctx, authClient, req, reqVerbs); err != nil {
			summary := fmt.Sprintf("missing permissions on resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(checkName, summary, err))
			continue
		}

		summary := fmt.Sprintf("%s: can %s", req.Resource, strings.Join(reqVerbs, ", "))
		results = append(results, api.PassResult(checkName, summary))
	}

	// TODO: delete this when the enterprise-helm role no longer requests resources on things
	// that don't exist.
	for req, reqVerbs := range k.reqs.RoleOnlyResourceRequirements {
		if err := k.checkOneRBACSSAR(ctx, authClient, req, reqVerbs); err != nil {
			summary := fmt.Sprintf("missing permissions on resource %s: %s", req.Resource, err)
			results = append(results, api.ErrorResult(checkName, summary, err))
			continue
		}

		summary := fmt.Sprintf("%s: can %s", req.Resource, strings.Join(reqVerbs, ", "))
		results = append(results, api.PassResult(checkName, summary))
	}

	return results
}

func (k *KubernetesChecker) checkOneRBACSSAR(ctx context.Context, authClient authorizationv1client.AuthorizationV1Interface, req *ResourceRequirement, reqVerbs ResourceVerbs) error {
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
