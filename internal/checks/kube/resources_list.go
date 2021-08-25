package kube

import (
	"github.com/Masterminds/semver/v3"

	"github.com/cdr/coder-doctor/internal/api"
)

var allVersionedRBACRequirements = []VersionedResourceRequirements{
	{
		VersionConstraints: api.MustConstraint(">= 1.20"),
		RBACRequirements: []*ResourceRequirement{
			NewResourceRequirement("", "v1", "pods", verbsCreateDeleteList...),
			NewResourceRequirement("", "v1", "secrets", verbsCreateDeleteList...),
			NewResourceRequirement("", "v1", "serviceaccounts", verbsCreateDeleteList...),
			NewResourceRequirement("", "v1", "services", verbsCreateDeleteList...),
			NewResourceRequirement("", "rbac.authorization.k8s.io/v1", "roles", verbsCreateDeleteList...),
			NewResourceRequirement("", "rbac.authorization.k8s.io/v1", "rolebindings", verbsCreateDeleteList...),
			NewResourceRequirement("apps", "apps/v1", "deployments", verbsCreateDeleteList...),
			NewResourceRequirement("apps", "apps/v1", "replicasets", verbsCreateDeleteList...),
			NewResourceRequirement("apps", "apps/v1", "statefulsets", verbsCreateDeleteList...),
			NewResourceRequirement("extensions", "ingresses", "networking.k8s.io/v1", verbsCreateDeleteList...),
		},
	},
}

// ResourceRequirement describes a set of requirements on a specific version of a resource:
// whether it exists with that specific version, and what verbs the current user is permitted to perform
// on the resource.
type ResourceRequirement struct {
	APIGroup string
	Resource string
	Verbs    []string
	Version  string
}

// VersionedResourceRequirements is a set of ResourceRequirements for a specific version of Coder.
type VersionedResourceRequirements struct {
	VersionConstraints *semver.Constraints
	RBACRequirements   []*ResourceRequirement
}

var verbsCreateDeleteList = []string{"create", "delete", "list"}

// NewResourceRequirement is just a convenience function for creating ResourceRequirements.uname
func NewResourceRequirement(apiGroup, version, resource string, verbs ...string) *ResourceRequirement {
	return &ResourceRequirement{
		APIGroup: apiGroup,
		Resource: resource,
		Verbs:    verbs,
		Version:  version,
	}
}
