package local

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"

	"cdr.dev/slog"
	"github.com/cdr/coder-doctor/internal/api"
)

const LocalHelmVersionCheck = "local-helm-version"

type VersionRequirement struct {
	Coder          *semver.Version
	HelmConstraint *semver.Constraints
}

var versionRequirements = []VersionRequirement{
	{
		Coder:          semver.MustParse("1.21.0"),
		HelmConstraint: api.MustConstraint(">= 3.6.0"),
	},
	{
		Coder:          semver.MustParse("1.20.0"),
		HelmConstraint: api.MustConstraint(">= 3.6.0"),
	},
}

func (l *Checker) CheckLocalHelmVersion(ctx context.Context) *api.CheckResult {
	if l.target != api.CheckTargetKubernetes {
		return api.SkippedResult(LocalHelmVersionCheck, "not applicable for target "+string(l.target))
	}

	helmBin, err := l.lookPathF("helm")
	if err != nil {
		return api.ErrorResult(LocalHelmVersionCheck, "could not find helm binary in $PATH", err)
	}

	helmVersionRaw, err := l.execF(ctx, helmBin, "version", "--short")
	if err != nil {
		return api.ErrorResult(LocalHelmVersionCheck, "failed to determine helm version", err)
	}

	helmVersion, err := semver.NewVersion(string(bytes.TrimSpace(helmVersionRaw)))
	if err != nil {
		return api.ErrorResult(LocalHelmVersionCheck, "failed to parse helm version", err)
	}

	selectedVersion := findNearestHelmVersion(l.coderVersion)
	l.log.Debug(ctx, "selected coder version", slog.F("requested", l.coderVersion), slog.F("selected", selectedVersion.Coder))

	result := &api.CheckResult{
		Name: LocalHelmVersionCheck,
		Details: map[string]interface{}{
			"helm-bin":                 helmBin,
			"helm-version":             helmVersion.String(),
			"helm-version-constraints": selectedVersion.HelmConstraint.String(),
		},
	}

	if ok, cerrs := selectedVersion.HelmConstraint.Validate(helmVersion); !ok {
		result.State = api.StateFailed
		var b strings.Builder
		_, err := fmt.Fprintf(&b, "Coder %s requires Helm version %s (installed: %s)\n", selectedVersion.Coder, selectedVersion.HelmConstraint, helmVersion)
		if err != nil {
			return api.ErrorResult(LocalHelmVersionCheck, "failed to write error result", err)
		}
		for _, cerr := range cerrs {
			if _, err := fmt.Fprintf(&b, "constraint failed: %s\n", cerr); err != nil {
				return api.ErrorResult(LocalHelmVersionCheck, "failed to write constraint error", err)
			}
		}
		result.Summary = b.String()
	} else {
		result.State = api.StatePassed
		result.Summary = fmt.Sprintf("Coder %s supports Helm %s", selectedVersion.Coder, selectedVersion.HelmConstraint)
	}

	return result
}

func findNearestHelmVersion(target *semver.Version) *VersionRequirement {
	var selected *VersionRequirement

	for _, v := range versionRequirements {
		v := v
		if !v.Coder.GreaterThan(target) {
			selected = &v
			break
		}
	}

	return selected
}
