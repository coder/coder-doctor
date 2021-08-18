package api

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"
)

type Checker interface {
	// Validate returns an error if, and only if, the Checker was not
	// configured correctly.
	//
	// This method is responsible for verifying that the Checker has
	// all required parameters and the required parameters are valid,
	// and that optional parameters are valid, if set.
	Validate() error

	// Run runs the checks and returns the results.
	//
	// This method will run through the checks and return results.
	Run(context.Context) error
}

var _ = fmt.Stringer(StatePassed)

type CheckState int

const (
	// StatePassed indicates that the check passed successfully.
	StatePassed CheckState = iota

	// StateWarning indicates a condition where Coder will gracefully degrade,
	// but the user will not have an optimal experience.
	StateWarning

	// StateFailed indicates a condition where Coder will not be able to install
	// successfully.
	StateFailed

	// StateInfo indicates a result for informational or diagnostic purposes
	// only, with no bearing on the ability to install Coder.
	StateInfo

	// StateSkipped indicates an indeterminate result due to a skipped check.
	StateSkipped
)

func (s CheckState) MustEmoji() string {
	emoji, err := s.Emoji()
	if err != nil {
		panic(err.Error())
	}
	return emoji
}

func (s CheckState) Emoji() (string, error) {
	switch s {
	case StatePassed:
		return "👍", nil
	case StateWarning:
		return "⚠️", nil
	case StateFailed:
		return "👎", nil
	case StateInfo:
		return "🔔", nil
	case StateSkipped:
		return "⏩", nil
	}

	return "", xerrors.Errorf("unknown state: %d", s)
}

func (s CheckState) MustText() string {
	text, err := s.Text()
	if err != nil {
		panic(err.Error())
	}
	return text
}

func (s CheckState) Text() (string, error) {
	switch s {
	case StatePassed:
		return "PASS", nil
	case StateWarning:
		return "WARN", nil
	case StateFailed:
		return "FAIL", nil
	case StateInfo:
		return "INFO", nil
	case StateSkipped:
		return "SKIP", nil
	}

	return "", xerrors.Errorf("unknown state: %d", s)
}

func (s CheckState) String() string {
	switch s {
	case StatePassed:
		return "StatePassed"
	case StateWarning:
		return "StateWarning"
	case StateFailed:
		return "StateFailed"
	case StateInfo:
		return "StateInfo"
	case StateSkipped:
		return "StateSkipped"
	}

	panic(fmt.Sprintf("unknown state: %d", s))
}

type CheckResult struct {
	Name    string                 `json:"name"`
	State   CheckState             `json:"state"`
	Summary string                 `json:"summary"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CheckTarget indicates the subject of a Checker
type CheckTarget string

const (
	// CheckTargetUndefined indicates that a Checker does not run against any specific target.
	CheckTargetUndefined CheckTarget = ""

	// CheckTargetLocal indicates that a Checker runs against the local machine.
	CheckTargetLocal CheckTarget = "local"

	// CheckTargetKubernetes indicates that a Checker runs against a Kubernetes cluster.
	CheckTargetKubernetes = "kubernetes"
)
