package kube

import (
	"context"
	"net/http"
	"testing"

	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"cdr.dev/slog/sloggers/slogtest/assert"

	"github.com/cdr/coder-doctor/internal/api"
)

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
					assert.True(t, result.Name+" should not error", result.Details["error"] == nil)
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

			server := newTestHTTPServer(t, http.StatusOK, test.Response)
			defer server.Close()

			client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
			assert.Success(t, "failed to create client", err)

			checker := NewKubernetesChecker(client)
			results := checker.checkRBACFallback(context.Background())
			test.F(t, results)
		})
	}
}

func Test_CheckRBACFallback_ClientError(t *testing.T) {
	t.Parallel()

	server := newTestHTTPServer(t, http.StatusInternalServerError, nil)

	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	assert.Success(t, "failed to create client", err)

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
	},
}

var selfSubjectAccessReviewDenied authorizationv1.SelfSubjectAccessReview = authorizationv1.SelfSubjectAccessReview{
	Status: authorizationv1.SubjectAccessReviewStatus{
		Allowed: false,
	},
}

func Test_CheckRBACDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name     string
		Response *authorizationv1.SelfSubjectRulesReview
		F        func(*testing.T, []*api.CheckResult)
	}{
		{
			Name:     "nothing allowed",
			Response: &selfSubjectRulesReviewEmpty,
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

			server := newTestHTTPServer(t, http.StatusOK, test.Response)
			defer server.Close()

			client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
			assert.Success(t, "failed to create client", err)

			checker := NewKubernetesChecker(client)
			results := checker.checkRBACDefault(context.Background())
			test.F(t, results)
		})
	}
}

func Test_Satisfies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
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
		// TODO(cian): add many, many, more.
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			actual := satisfies(testCase.Requirement, testCase.Verbs, testCase.Rules)
			if testCase.Expected == nil {
				assert.Success(t, testCase.Name+"- should not error", actual)
			} else {
				assert.ErrorContains(t, testCase.Name+" - expected error contains", actual, *testCase.Expected)
			}
		})
	}
}

var testNoPermRule = makeTestPolicyRule(ss(), ss(), ss(), ss(), ss())
var testV1WildcardRule = makeTestPolicyRule(verbsAll, ss(""), ss("*"), ss(), ss())

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
