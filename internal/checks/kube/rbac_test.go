package kube

import (
	"context"
	"testing"

	"golang.org/x/xerrors"
	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"

	"cdr.dev/slog/sloggers/slogtest/assert"

	"github.com/cdr/coder-doctor/internal/api"
)

func Test_CheckRBAC_Error(t *testing.T) {
	t.Parallel()
	srv := newTestHTTPServer(t, 500, nil)
	defer srv.Close()
	client, err := kubernetes.NewForConfig(&rest.Config{Host: srv.URL})
	assert.Success(t, "failed to create client", err)

	checker := NewKubernetesChecker(client)
	results := checker.CheckRBAC(context.Background())
	assert.True(t, "should contain one result", len(results) == 1)
	assert.True(t, "result should be failed", results[0].State == api.StateFailed)
}

func Test_CheckRBACFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name     string
		Response *authorizationv1.SelfSubjectAccessReview
		F        func(*testing.T, []*api.CheckResult)
	}{
		{
			Name:     "all allowed",
			Response: &selfSubjectAccessReviewAllowed,
			F: func(t *testing.T, results []*api.CheckResult) {
				assert.False(t, "results should not be empty", len(results) == 0)
				for _, result := range results {
					assert.Equal(t, result.Name+" should not error", result.Details["error"], nil)
					assert.True(t, result.Name+" should pass", result.State == api.StatePassed)
				}
			},
		},
		{
			Name:     "all denied",
			Response: &selfSubjectAccessReviewDenied,
			F: func(t *testing.T, results []*api.CheckResult) {
				assert.False(t, "results should not be empty", len(results) == 0)
				for _, result := range results {
					assert.True(t, result.Name+" should have an error", result.Details["error"] != nil)
					assert.True(t, result.Name+" should fail", result.State == api.StateFailed)
				}
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			client := fake.NewSimpleClientset()
			fakeAction := func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, test.Response, nil
			}
			// NOTE: Use PrependReactor! AddReactor appends the action after the reaction chain
			// which by default includes a "catch-all" action which is not what we want here!
			client.Fake.PrependReactor("create", "selfsubjectaccessreviews", fakeAction)

			checker := NewKubernetesChecker(client)
			results := checker.checkRBACFallback(context.Background())
			test.F(t, results)
		})
	}
}

func Test_CheckRBACFallback_ClientError(t *testing.T) {
	t.Parallel()
	client := fake.NewSimpleClientset()
	fakeAction := func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, xerrors.New("ouch")
	}
	// NOTE: Use PrependReactor! AddReactor appends the action after the reaction chain
	// which by default includes a "catch-all" action which is not what we want here!
	client.Fake.PrependReactor("create", "selfsubjectaccessreviews", fakeAction)

	checker := NewKubernetesChecker(client)
	results := checker.checkRBACFallback(context.Background())
	for _, result := range results {
		assert.ErrorContains(t, result.Name+" should show correct error", result.Details["error"].(error), "failed to create SelfSubjectAccessReview request")
		assert.True(t, result.Name+" should fail", result.State == api.StateFailed)
	}
}

var selfSubjectAccessReviewAllowed authorizationv1.SelfSubjectAccessReview = authorizationv1.SelfSubjectAccessReview{
	Status: authorizationv1.SubjectAccessReviewStatus{
		Allowed: true,
		Reason:  "test says yes",
	},
}

var selfSubjectAccessReviewDenied authorizationv1.SelfSubjectAccessReview = authorizationv1.SelfSubjectAccessReview{
	Status: authorizationv1.SubjectAccessReviewStatus{
		Allowed: false,
		Reason:  "test says no",
	},
}

func Test_CheckRBACDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name     string
		Response *authorizationv1.SelfSubjectRulesReview
		TestFunc func(*testing.T, []*api.CheckResult, error)
	}{
		{
			Name:     "nothing allowed",
			Response: &selfSubjectRulesReviewEmpty,
			TestFunc: func(t *testing.T, results []*api.CheckResult, err error) {
				assert.False(t, "results should not be empty", len(results) == 0)
				for _, result := range results {
					assert.True(t, result.Name+" should not return an error", err == nil)
					assert.True(t, result.Name+" should contain an error in details", result.Details["error"] != nil)
					assert.True(t, result.Name+" should fail", result.State == api.StateFailed)
				}
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			client := fake.NewSimpleClientset()

			fakeAction := func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, test.Response, nil
			}
			// NOTE: Use PrependReactor! AddReactor appends the action after the reaction chain
			// which by default includes a "catch-all" action which is not what we want here!
			client.Fake.PrependReactor("create", "selfsubjectrulesreviews", fakeAction)

			checker := NewKubernetesChecker(client)
			results, err := checker.checkRBACDefault(context.Background())
			test.TestFunc(t, results, err)
		})
	}
}

func Test_Satisfies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name        string
		Requirement *ResourceRequirement
		Verbs       ResourceVerbs
		Rules       []rbacv1.PolicyRule
		Expected    *string
	}{
		{
			Name:        "allowed: get create pods",
			Requirement: NewResourceRequirement("", "v1", "pods"),
			Verbs:       verbsGetCreate,
			Rules:       []rbacv1.PolicyRule{testV1WildcardRule},
			Expected:    nil,
		},
		{
			Name:        "not allowed: get create pods",
			Requirement: NewResourceRequirement("", "v1", "pods"),
			Verbs:       verbsGetCreate,
			Rules:       []rbacv1.PolicyRule{testNoPermRule},
			Expected:    strptr("not satisfied"),
		},
		{
			Name:        "not allowed: get create apps/deployments",
			Requirement: NewResourceRequirement("apps", "apps/v1", "deployments"),
			Verbs:       verbsGetCreate,
			Rules:       []rbacv1.PolicyRule{testV1WildcardRule},
			Expected:    strptr("not satisfied"),
		},
		{
			Name:        "allowed: get create apps/deployments",
			Requirement: NewResourceRequirement("apps", "apps/v1", "deployments"),
			Verbs:       verbsGetCreate,
			Rules:       []rbacv1.PolicyRule{testAppsV1WildcardRule},
			Expected:    nil,
		},
		{
			Name:        "not allowed: delete apps/deployments",
			Requirement: NewResourceRequirement("apps", "apps/v1", "deployments"),
			Verbs:       ss("delete"),
			Rules: []rbacv1.PolicyRule{
				makeTestPolicyRule(verbsGetCreate, ss("apps"), ss("*"), ss(), ss()),
			},
			Expected: strptr("not satisfied"),
		},
		{
			Name:        "allowed: verb wildcard",
			Requirement: NewResourceRequirement("", "v1", "pods"),
			Verbs:       verbsGetCreate,
			Rules: []rbacv1.PolicyRule{
				makeTestPolicyRule(ss("*"), ss(""), ss("pods"), ss(), ss()),
			},
			Expected: nil,
		},
		{
			Name:        "allowed: resource wildcard",
			Requirement: NewResourceRequirement("", "v1", "pods"),
			Verbs:       verbsGetCreate,
			Rules: []rbacv1.PolicyRule{
				makeTestPolicyRule(verbsGetCreate, ss(""), ss("*"), ss(), ss()),
			},
			Expected: nil,
		},
		{
			Name:        "allowed: group wildcard",
			Requirement: NewResourceRequirement("apps", "apps/v1", "deployments"),
			Verbs:       verbsGetCreate,
			Rules: []rbacv1.PolicyRule{
				makeTestPolicyRule(verbsGetCreate, ss("*"), ss("*"), ss(), ss()),
			},
			Expected: nil,
		},
		// TODO(cian): add many, many, more.
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			actual := satisfies(test.Requirement, test.Verbs, test.Rules)
			if test.Expected == nil {
				assert.Success(t, test.Name+"- should not error", actual)
			} else {
				assert.ErrorContains(t, test.Name+" - expected error contains", actual, *test.Expected)
			}
		})
	}
}

var testNoPermRule = makeTestPolicyRule(ss(), ss(), ss(), ss(), ss())
var testV1WildcardRule = makeTestPolicyRule(verbsAll, ss(""), ss("*"), ss(), ss())
var testAppsV1WildcardRule = makeTestPolicyRule(verbsAll, ss("apps"), ss("*"), ss(), ss())

func makeTestPolicyRule(verbs, groups, resources, resourceNames, nonResourceURLs []string) rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		Verbs:           verbs,
		APIGroups:       groups,
		Resources:       resources,
		ResourceNames:   resourceNames,
		NonResourceURLs: nonResourceURLs,
	}
}

var selfSubjectRulesReviewEmpty = authorizationv1.SelfSubjectRulesReview{
	Spec: authorizationv1.SelfSubjectRulesReviewSpec{
		Namespace: "default",
	},
	Status: authorizationv1.SubjectRulesReviewStatus{
		ResourceRules:    []authorizationv1.ResourceRule{},
		NonResourceRules: []authorizationv1.NonResourceRule{},
	},
}

func strptr(s string) *string {
	return &s
}

func ss(s ...string) []string {
	return s
}
