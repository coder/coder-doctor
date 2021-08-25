package kube

import (
	"context"
	"net/http"
	"testing"

	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"cdr.dev/slog/sloggers/slogtest/assert"

	"github.com/cdr/coder-doctor/internal/api"
)

func Test_CheckRBAC(t *testing.T) {
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
			results := checker.CheckRBAC(context.Background())
			test.F(t, results)
		})
	}
}

func Test_CheckRBAC_ClientError(t *testing.T) {
	t.Parallel()

	server := newTestHTTPServer(t, http.StatusInternalServerError, nil)

	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	assert.Success(t, "failed to create client", err)

	checker := NewKubernetesChecker(client)
	results := checker.CheckRBAC(context.Background())
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
