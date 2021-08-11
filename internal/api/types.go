package api

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"k8s.io/client-go/kubernetes"
)

type Check func(context.Context, CheckOptions) CheckResults

type CheckOptions struct {
	CoderVersion *semver.Version
	Kubernetes   kubernetes.Interface
}

type CheckState int

const (
	StatePassed CheckState = iota
	StateWarning
	StateFailed
	StateInfo
	StateSkipped
)

type CheckResult struct {
	Name    string
	State   CheckState
	Summary string
	Details map[string]interface{}
}

func ErrorResult(name string, summary string, err error) *CheckResult {
	return &CheckResult{
		Name:    name,
		State:   StateFailed,
		Summary: summary,
		Details: map[string]interface{}{
			"error": err,
		},
	}
}

type CheckResults []*CheckResult
