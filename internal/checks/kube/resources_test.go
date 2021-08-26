package kube

import (
	"context"
	"net/http"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/cdr/coder-doctor/internal/api"
)

func Test_KubernetesChecker_CheckResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name     string
		Response *metav1.APIResourceList
		F        func(*testing.T, []*api.CheckResult)
	}{
		{
			Name:     "no resources available",
			Response: emptyAPIResourceList,
			F: func(t *testing.T, results []*api.CheckResult) {
				assert.False(t, "results should not be empty", len(results) == 0)
				for _, result := range results {
					resErr, ok := result.Details["error"].(error)
					assert.True(t, result.Name+" should have an error", ok && resErr != nil)
					assert.ErrorContains(t, result.Name+" should have an expected error", resErr, "missing required resource")
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
			results := checker.CheckResources(context.Background())
			test.F(t, results)
		})
	}
}

var emptyAPIResourceList *metav1.APIResourceList = &metav1.APIResourceList{
	TypeMeta:     metav1.TypeMeta{},
	GroupVersion: "",
	APIResources: []metav1.APIResource{},
}
