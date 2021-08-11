package kubernetes_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cdr.dev/slog/sloggers/slogtest/assert"
	"github.com/Masterminds/semver/v3"
	"github.com/cdr/coder-doctor/internal/api"
	"github.com/cdr/coder-doctor/internal/checks/kubernetes"
	"k8s.io/apimachinery/pkg/version"
	kclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name              string
		CoderVersion      *semver.Version
		KubernetesVersion *version.Info
		ExpectedResult    *api.CheckResult
	}{
		{
			Name:         "coder-valid-version-gke",
			CoderVersion: semver.MustParse("1.21"),
			KubernetesVersion: &version.Info{
				Major:        "1",
				Minor:        "20+",
				GitVersion:   "v1.20.8-gke.900",
				GitCommit:    "28ab8501be88ea42e897ca8514d7cd0b436253d9",
				GitTreeState: "clean",
				BuildDate:    "2021-06-30T09:23:36Z",
				GoVersion:    "go1.15.13b5",
				Compiler:     "gc",
				Platform:     "linux/amd64",
			},
			ExpectedResult: &api.CheckResult{
				Name:    "kubernetes-version",
				State:   api.StatePassed,
				Summary: "Coder 1.21.0 supports Kubernetes 1.19.0 to 1.22.0 (server version 1.20.8-gke.900)",
				Details: map[string]interface{}{
					"build-date":     "2021-06-30T09:23:36Z",
					"compiler":       "gc",
					"git-commit":     "28ab8501be88ea42e897ca8514d7cd0b436253d9",
					"git-tree-state": "clean",
					"git-version":    "v1.20.8-gke.900",
					"go-version":     "go1.15.13b5",
					"major":          "1",
					"minor":          "20+",
					"platform":       "linux/amd64",
				},
			},
		},
		{
			Name:         "coder-old-version-gke",
			CoderVersion: semver.MustParse("1.21"),
			KubernetesVersion: &version.Info{
				Major:        "1",
				Minor:        "18+",
				GitVersion:   "v1.18.20-gke.900",
				GitCommit:    "1facb91642e16cb4f5be4e4a632c488aa4700382",
				GitTreeState: "clean",
				BuildDate:    "2021-06-28T09:19:58Z",
				GoVersion:    "go1.13.15b4",
				Compiler:     "gc",
				Platform:     "linux/amd64",
			},
			ExpectedResult: &api.CheckResult{
				Name:    "kubernetes-version",
				State:   2,
				Summary: "Coder 1.21.0 supports Kubernetes 1.19.0 to 1.22.0 and was not tested with 1.18.20-gke.900",
				Details: map[string]interface{}{
					"build-date":     "2021-06-28T09:19:58Z",
					"compiler":       "gc",
					"git-commit":     "1facb91642e16cb4f5be4e4a632c488aa4700382",
					"git-tree-state": "clean",
					"git-version":    "v1.18.20-gke.900",
					"go-version":     "go1.13.15b4",
					"major":          "1",
					"minor":          "18+",
					"platform":       "linux/amd64",
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(test.KubernetesVersion)
				assert.Success(t, "failed to encode response", err)
			}))
			defer server.Close()

			client, err := kclient.NewForConfig(&rest.Config{
				Host: server.URL,
			})
			assert.Success(t, "failed to create client", err)

			res := kubernetes.CheckVersion(context.Background(), api.CheckOptions{
				CoderVersion: test.CoderVersion,
				Kubernetes:   client,
			})
			assert.Equal(t, "check result matches", test.ExpectedResult, res)
		})
	}
}
