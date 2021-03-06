package kube

import (
	"context"
	"net/http"
	"testing"

	"github.com/Masterminds/semver/v3"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"cdr.dev/coder-doctor/internal/api"
	"cdr.dev/slog/sloggers/slogtest/assert"
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
					"coder-version":       "1.21.0",
					"coder-version-major": uint64(1),
					"coder-version-minor": uint64(21),
					"coder-version-patch": uint64(0),
					"build-date":          "2021-06-30T09:23:36Z",
					"compiler":            "gc",
					"git-commit":          "28ab8501be88ea42e897ca8514d7cd0b436253d9",
					"git-tree-state":      "clean",
					"git-version":         "v1.20.8-gke.900",
					"go-version":          "go1.15.13b5",
					"major":               "1",
					"minor":               "20+",
					"platform":            "linux/amd64",
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
				State:   api.StateFailed,
				Summary: "Coder 1.21.0 supports Kubernetes 1.19.0 to 1.22.0 and was not tested with 1.18.20-gke.900",
				Details: map[string]interface{}{
					"coder-version":       "1.21.0",
					"coder-version-major": uint64(1),
					"coder-version-minor": uint64(21),
					"coder-version-patch": uint64(0),
					"build-date":          "2021-06-28T09:19:58Z",
					"compiler":            "gc",
					"git-commit":          "1facb91642e16cb4f5be4e4a632c488aa4700382",
					"git-tree-state":      "clean",
					"git-version":         "v1.18.20-gke.900",
					"go-version":          "go1.13.15b4",
					"major":               "1",
					"minor":               "18+",
					"platform":            "linux/amd64",
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			server := newTestHTTPServer(t, http.StatusOK, test.KubernetesVersion)
			defer server.Close()

			client, err := kubernetes.NewForConfig(&rest.Config{
				Host: server.URL,
			})
			assert.Success(t, "failed to create client", err)

			checker := NewKubernetesChecker(client, WithCoderVersion(test.CoderVersion))
			result := checker.CheckVersion(context.Background())
			assert.Equal(t, "check result matches", test.ExpectedResult, result)
		})
	}
}

func TestUnknownRoute(t *testing.T) {
	t.Parallel()

	server := newTestHTTPServer(t, http.StatusNotFound, nil)
	defer server.Close()

	client, err := kubernetes.NewForConfig(&rest.Config{
		Host: server.URL,
	})
	assert.Success(t, "failed to create client", err)

	checker := NewKubernetesChecker(client)
	result := checker.CheckVersion(context.Background())
	assert.Equal(t, "failed check", api.StateFailed, result.State)
	assert.ErrorContains(t, "unknown route", result.Details["error"].(error), "the server could not find the requested resource")
}

func TestCorruptResponse(t *testing.T) {
	t.Parallel()

	garbage := map[string]interface{}{"gitVersion": 10}
	server := newTestHTTPServer(t, http.StatusOK, garbage)
	defer server.Close()

	client, err := kubernetes.NewForConfig(&rest.Config{
		Host: server.URL,
	})
	assert.Success(t, "failed to create client", err)

	checker := NewKubernetesChecker(client)
	result := checker.CheckVersion(context.Background())
	assert.Equal(t, "failed check", api.StateFailed, result.State)
	assert.ErrorContains(t, "unknown route", result.Details["error"].(error), "json: cannot unmarshal number into Go struct field Info.gitVersion of type string")
}
func TestNearestVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name             string
		RequestedVersion string
		NearestVersion   string
	}{
		{
			Name:             "exact-match",
			RequestedVersion: "1.20.0",
			NearestVersion:   "1.20.0",
		},
		{
			Name:             "nearby-match",
			RequestedVersion: "1.20.1",
			NearestVersion:   "1.20.0",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			requestedVersion := semver.MustParse(test.RequestedVersion)
			nearestVersion := semver.MustParse(test.NearestVersion)

			found := findNearestVersion(requestedVersion)
			assert.Equal(t, "nearest version matches", nearestVersion, found.CoderVersion)
		})
	}
}
