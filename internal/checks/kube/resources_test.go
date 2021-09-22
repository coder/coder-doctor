package kube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"cdr.dev/coder-doctor/internal/api"
	"cdr.dev/slog/sloggers/slogtest/assert"
)

func Test_KubernetesChecker_CheckResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name              string
		GroupListResponse *metav1.APIGroupList
		F                 func(*testing.T, []*api.CheckResult)
	}{
		{
			Name:              "all resources available",
			GroupListResponse: fullAPIGroupList,
			F: func(t *testing.T, results []*api.CheckResult) {
				assert.False(t, "results should not be empty", len(results) == 0)
				for _, result := range results {
					assert.Equal(t, result.Name+" should have no error", nil, result.Details["error"])
					assert.Equal(t, result.Name+" should pass", api.StatePassed, result.State)
				}
			},
		},
		{
			Name:              "no resources available",
			GroupListResponse: emptyAPIGroupList,
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

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/apis":
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(test.GroupListResponse)
					assert.Success(t, "failed to encode response", err)
				case "/api/v1":
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(v1ResourceList)
					assert.Success(t, "failed to encode response", err)
				case "/apis/apps/v1":
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(appsV1ResourceList)
					assert.Success(t, "failed to encode response", err)
				case "/apis/metrics.k8s.io/v1":
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(metricsV1ResourceLilst)
					assert.Success(t, "failed to encode response", err)
				case "/apis/networking.k8s.io/v1":
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(networkingV1ResourceList)
					assert.Success(t, "failed to encode response", err)
				case "/apis/rbac.authorization.k8s.io/v1":
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(rbacV1ResourceList)
					assert.Success(t, "failed to encode response", err)
				case "/apis/storage.k8s.io/v1":
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(storageV1ResourceList)
					assert.Success(t, "failed to encode response", err)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
			assert.Success(t, "failed to create client", err)

			checker := NewKubernetesChecker(client)
			results := checker.CheckResources(context.Background())
			test.F(t, results)
		})
	}
}

var emptyAPIGroupList *metav1.APIGroupList = &metav1.APIGroupList{
	TypeMeta: metav1.TypeMeta{
		Kind:       "APIGroupList",
		APIVersion: "v1",
	},
	Groups: []metav1.APIGroup{},
}

var fullAPIGroupList *metav1.APIGroupList = &metav1.APIGroupList{
	TypeMeta: metav1.TypeMeta{
		Kind:       "APIGroupList",
		APIVersion: "v1",
	},
	Groups: []metav1.APIGroup{
		{
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: "v1",
					Version:      "v1",
				},
			},
		},
		{
			Name: "apps",
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: "apps/v1",
					Version:      "v1",
				},
			},
		},
		{
			Name: "metrics.k8s.io",
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: "metrics.k8s.io/v1beta1",
					Version:      "v1",
				},
			},
		},
		{
			Name: "networking.k8s.io",
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: "networking.k8s.io/v1",
					Version:      "v1",
				},
			},
		},
		{
			Name: "rbac.authorization.k8s.io",
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: "rbac.authorization.k8s.io/v1",
					Version:      "v1",
				},
			},
		},
		{
			Name: "storage.k8s.io",
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: "storage.k8s.io/v1",
					Version:      "v1",
				},
			},
		},
	},
}

var v1ResourceList = makeResourceList("", "v1", "persistentvolumes", "persistentvolumeclaims", "events", "pods", "secrets", "serviceaccounts", "services")
var appsV1ResourceList = makeResourceList("v1", "apps/v1", "deployments", "replicasets", "statefulsets")
var metricsV1ResourceLilst = makeResourceList("v1", "metrics.k8s.io/v1beta1", "pods")
var networkingV1ResourceList = makeResourceList("v1", "networking.k8s.io/v1", "ingresses", "networkpolicies")
var rbacV1ResourceList = makeResourceList("v1", "rbac.authorization.k8s.io/v1", "roles", "rolebindings")
var storageV1ResourceList = makeResourceList("v1", "storage.k8s.io/v1", "roles", "storageclasses")

func makeResourceList(version, groupVersion string, resources ...string) *metav1.APIResourceList {
	rl := &metav1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResourceList",
			APIVersion: version,
		},
		GroupVersion: groupVersion,
		APIResources: []metav1.APIResource{},
	}
	for _, resource := range resources {
		rl.APIResources = append(rl.APIResources, metav1.APIResource{Name: resource, Verbs: []string{"get"}})
	}
	return rl
}
