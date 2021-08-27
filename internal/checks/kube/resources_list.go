package kube

import (
	"github.com/Masterminds/semver/v3"

	"github.com/cdr/coder-doctor/internal/api"
)

var allRequirements = []VersionedResourceRequirements{
	{
		VersionConstraints: api.MustConstraint(">= 1.20"),
		ResourceRequirements: map[*ResourceRequirement]ResourceVerbs{
			NewResourceRequirement("", "v1", "pods"):                                                            verbsCreateDeleteList,
			NewResourceRequirement("", "v1", "secrets"):                                                         verbsCreateDeleteList,
			NewResourceRequirement("", "v1", "serviceaccounts"):                                                 verbsCreateDeleteList,
			NewResourceRequirement("", "v1", "services"):                                                        verbsCreateDeleteList,
			NewResourceRequirement("apps", "apps/v1", "deployments"):                                            verbsCreateDeleteList,
			NewResourceRequirement("apps", "apps/v1", "replicasets"):                                            verbsCreateDeleteList,
			NewResourceRequirement("apps", "apps/v1", "statefulsets"):                                           verbsCreateDeleteList,
			NewResourceRequirement("networking.k8s.io", "networking.k8s.io/v1", "ingresses"):                    verbsCreateDeleteList,
			NewResourceRequirement("rbac.authorization.k8s.io", "rbac.authorization.k8s.io/v1", "roles"):        verbsCreateDeleteList,
			NewResourceRequirement("rbac.authorization.k8s.io", "rbac.authorization.k8s.io/v1", "rolebindings"): verbsCreateDeleteList,
		},
	},
}

// ResourceRequirement describes a set of requirements on a specific version of a resource:
// whether it exists with that specific version, and what verbs the current user is permitted to perform
// on the resource.
type ResourceRequirement struct {
	Group    string
	Resource string
	Version  string
}

type ResourceVerbs []string

// VersionedResourceRequirements is a set of ResourceRequirements for a specific version of Coder.
type VersionedResourceRequirements struct {
	VersionConstraints   *semver.Constraints
	ResourceRequirements map[*ResourceRequirement]ResourceVerbs
}

var verbsCreateDeleteList ResourceVerbs = []string{"create", "delete", "list"}

// NewResourceRequirement is just a convenience function for creating ResourceRequirements.
func NewResourceRequirement(apiGroup, version, resource string) *ResourceRequirement {
	return &ResourceRequirement{
		Group:    apiGroup,
		Resource: resource,
		Version:  version,
	}
}
