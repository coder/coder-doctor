package api

import (
	"github.com/Masterminds/semver/v3"
	"golang.org/x/xerrors"
)

func MustConstraint(s string) *semver.Constraints {
	c, err := semver.NewConstraint(s)
	if err != nil {
		panic(xerrors.Errorf("parse constraint: %w", err))
	}

	return c
}
