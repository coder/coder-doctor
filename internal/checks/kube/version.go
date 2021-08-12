package kube

import (
	"context"
	"encoding/json"
	"fmt"

	"cdr.dev/slog"
	"github.com/Masterminds/semver/v3"
	"k8s.io/apimachinery/pkg/version"

	"github.com/cdr/coder-doctor/internal/api"
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

func findNearestVersion(coderVersion *semver.Version) *CoderVersionRequirement {
	var selectedVersion *CoderVersionRequirement

	for _, v := range versionRequirements {
		v := v
		if !v.CoderVersion.GreaterThan(coderVersion) {
			selectedVersion = &v
			break
		}
	}

	return selectedVersion
}

func (k *KubernetesChecker) CheckVersion(ctx context.Context) *api.CheckResult {
	const checkName = "kubernetes-version"

	var versionInfo version.Info

	// This uses the RESTClient rather than Discovery().ServerVersion()
	// because the latter does not accept a context.
	body, err := k.client.Discovery().RESTClient().Get().AbsPath("/version").Do(ctx).Raw()
	if err != nil {
		return api.ErrorResult(checkName, "failed to get version from server", err)
	}

	err = json.Unmarshal(body, &versionInfo)
	if err != nil {
		return api.ErrorResult(checkName, "failed to parse server version", err)
	}

	selectedVersion := findNearestVersion(k.coderVersion)
	k.log.Debug(ctx, "selected coder version",
		slog.F("requested", k.coderVersion),
		slog.F("selected", selectedVersion.CoderVersion))

	kubernetesVersion, err := semver.NewVersion(versionInfo.GitVersion)
	if err != nil {
		return api.ErrorResult(checkName, "failed to parse server version", err)
	}

	result := &api.CheckResult{
		Name: checkName,
		Details: map[string]interface{}{
			"coder-version":       selectedVersion.CoderVersion.String(),
			"coder-version-major": selectedVersion.CoderVersion.Major(),
			"coder-version-minor": selectedVersion.CoderVersion.Minor(),
			"coder-version-patch": selectedVersion.CoderVersion.Patch(),
			"platform":            versionInfo.Platform,
			"major":               versionInfo.Major,
			"minor":               versionInfo.Minor,
			"git-version":         versionInfo.GitVersion,
			"git-commit":          versionInfo.GitCommit,
			"git-tree-state":      versionInfo.GitTreeState,
			"build-date":          versionInfo.BuildDate,
			"go-version":          versionInfo.GoVersion,
			"compiler":            versionInfo.Compiler,
		},
	}

	if kubernetesVersion.LessThan(selectedVersion.KubernetesVersionMin) || kubernetesVersion.GreaterThan(selectedVersion.KubernetesVersionMax) {
		result.State = api.StateFailed
		result.Summary = fmt.Sprintf("Coder %s supports Kubernetes %s to %s and was not tested with %s",
			k.coderVersion, selectedVersion.KubernetesVersionMin, selectedVersion.KubernetesVersionMax, kubernetesVersion)
	} else {
		result.State = api.StatePassed
		result.Summary = fmt.Sprintf("Coder %s supports Kubernetes %s to %s (server version %s)",
			k.coderVersion, selectedVersion.KubernetesVersionMin, selectedVersion.KubernetesVersionMax, kubernetesVersion)
	}

	return result
}
