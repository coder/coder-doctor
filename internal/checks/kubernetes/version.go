package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/cdr/coder-doctor/internal/api"
	"k8s.io/apimachinery/pkg/version"
)

type CoderVersionRequirement struct {
	CoderVersion         *semver.Version
	KubernetesVersionMin *semver.Version
	KubernetesVersionMax *semver.Version
}

var versionRequirements = []CoderVersionRequirement{
	{
		CoderVersion:         semver.MustParse("1.21"),
		KubernetesVersionMin: semver.MustParse("1.19"),
		KubernetesVersionMax: semver.MustParse("1.22"),
	}, {
		CoderVersion:         semver.MustParse("1.20"),
		KubernetesVersionMin: semver.MustParse("1.19"),
		KubernetesVersionMax: semver.MustParse("1.21"),
	},
}

func CheckVersion(ctx context.Context, opts api.CheckOptions) *api.CheckResult {
	const checkName = "kubernetes-version"

	coderVersion := opts.CoderVersion
	client := opts.Kubernetes

	var versionInfo version.Info

	// This uses the RESTClient rather than Discovery().ServerVersion()
	// because the latter does not accept a context.
	body, err := client.Discovery().RESTClient().Get().AbsPath("/version").Do(ctx).Raw()
	if err != nil {
		return api.ErrorResult(checkName, "failed to get version from server", err)
	}

	err = json.Unmarshal(body, &versionInfo)
	if err != nil {
		return api.ErrorResult(checkName, "failed to parse server version", err)
	}

	var v CoderVersionRequirement
	for _, v = range versionRequirements {
		if !v.CoderVersion.LessThan(coderVersion) {
			break
		}
	}

	kubernetesVersion, err := semver.NewVersion(strings.TrimLeft(versionInfo.GitVersion, "v"))
	if err != nil {
		fmt.Printf("error parsing version: %v\n", err)
	}

	result := &api.CheckResult{
		Name: checkName,
		Details: map[string]interface{}{
			"platform":       versionInfo.Platform,
			"major":          versionInfo.Major,
			"minor":          versionInfo.Minor,
			"git-version":    versionInfo.GitVersion,
			"git-commit":     versionInfo.GitCommit,
			"git-tree-state": versionInfo.GitTreeState,
			"build-date":     versionInfo.BuildDate,
			"go-version":     versionInfo.GoVersion,
			"compiler":       versionInfo.Compiler,
		},
	}

	if kubernetesVersion.LessThan(v.KubernetesVersionMin) || kubernetesVersion.GreaterThan(v.KubernetesVersionMax) {
		result.State = api.StateFailed
		result.Summary = fmt.Sprintf("Coder %s supports Kubernetes %s to %s and was not tested with %s",
			v.CoderVersion, v.KubernetesVersionMin, v.KubernetesVersionMax, kubernetesVersion)
	} else {
		result.State = api.StatePassed
		result.Summary = fmt.Sprintf("Coder %s supports Kubernetes %s to %s (server version %s)",
			v.CoderVersion, v.KubernetesVersionMin, v.KubernetesVersionMax, kubernetesVersion)
	}

	// fmt.Printf("server version: %v\n", kubernetesVersion)
	// fmt.Printf("min version: %v\n", v.KubernetesVersionMin)
	// fmt.Printf("max version: %v\n", v.KubernetesVersionMax)

	return result
}
