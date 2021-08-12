package api

import (
	"context"
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
	Run(context.Context) CheckResults
}

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

type CheckResult struct {
	Name    string
	State   CheckState
	Summary string
	Details map[string]interface{}
}

type CheckResults []*CheckResult
