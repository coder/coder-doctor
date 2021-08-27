package kube

import (
	"github.com/Masterminds/semver/v3"

	"github.com/cdr/coder-doctor/internal/api"
)

// This is a list of all known Resource and/or RBAC requirements for each verison of Coder.
// Order by version DESCENDING.
var allRequirements = []VersionedResourceRequirements{
	{
		VersionConstraints: api.MustConstraint(">= 1.20"),
		ResourceRequirements: map[*ResourceRequirement]ResourceVerbs{
			NewResourceRequirement("", "v1", "events"):                                                          verbsAll,
			NewResourceRequirement("", "v1", "persistentvolumeclaims"):                                          verbsAll,
			NewResourceRequirement("", "v1", "pods"):                                                            verbsAll,
			NewResourceRequirement("", "v1", "secrets"):                                                         verbsAll,
			NewResourceRequirement("", "v1", "serviceaccounts"):                                                 verbsAll,
			NewResourceRequirement("", "v1", "services"):                                                        verbsAll,
			NewResourceRequirement("apps", "apps/v1", "deployments"):                                            verbsAll,
			NewResourceRequirement("apps", "apps/v1", "replicasets"):                                            verbsAll,
			NewResourceRequirement("apps", "apps/v1", "statefulsets"):                                           verbsAll,
			NewResourceRequirement("metrics.k8s.io", "metrics.k8s.io/v1beta1", "pods"):                          verbsGetListWatch,
			NewResourceRequirement("networking.k8s.io", "networking.k8s.io/v1", "ingresses"):                    verbsAll,
			NewResourceRequirement("networking.k8s.io", "networking.k8s.io/v1", "networkpolicies"):              verbsAll,
			NewResourceRequirement("rbac.authorization.k8s.io", "rbac.authorization.k8s.io/v1", "roles"):        verbsGetCreate,
			NewResourceRequirement("rbac.authorization.k8s.io", "rbac.authorization.k8s.io/v1", "rolebindings"): verbsGetCreate,
			NewResourceRequirement("storage.k8s.io", "storage.k8s.io/v1", "storageclasses"):                     verbsGetListWatch,
		},
		RoleOnlyResourceRequirements: map[*ResourceRequirement]ResourceVerbs{
			// The below permissions are required by the default coder role created by the Helm chart.
			// Installation will fail if these are not present in the role being used to install Coder.
			NewResourceRequirement("", "v1", "deployments"):                             verbsAll,
			NewResourceRequirement("", "v1", "networkpolicies"):                         verbsAll,
			NewResourceRequirement("", "v1", "pods/exec"):                               verbsAll,
			NewResourceRequirement("", "v1", "pods/log"):                                verbsAll,
			NewResourceRequirement("apps", "v1", "events"):                              verbsAll,
			NewResourceRequirement("apps", "v1", "networkpolicies"):                     verbsAll,
			NewResourceRequirement("apps", "v1", "persistentvolumeclaims"):              verbsAll,
			NewResourceRequirement("apps", "v1", "pods"):                                verbsAll,
			NewResourceRequirement("apps", "v1", "pods/exec"):                           verbsAll,
			NewResourceRequirement("apps", "v1", "pods/log"):                            verbsAll,
			NewResourceRequirement("apps", "v1", "secrets"):                             verbsAll,
			NewResourceRequirement("apps", "v1", "services"):                            verbsAll,
			NewResourceRequirement("metrics.k8s.io", "v1beta1", "storageclasses"):       verbsGetListWatch,
			NewResourceRequirement("networking.k8s.io", "v1", "deployments"):            verbsAll,
			NewResourceRequirement("networking.k8s.io", "v1", "events"):                 verbsAll,
			NewResourceRequirement("networking.k8s.io", "v1", "persistentvolumeclaims"): verbsAll,
			NewResourceRequirement("networking.k8s.io", "v1", "pods"):                   verbsAll,
			NewResourceRequirement("networking.k8s.io", "v1", "pods/exec"):              verbsAll,
			NewResourceRequirement("networking.k8s.io", "v1", "pods/log"):               verbsAll,
			NewResourceRequirement("networking.k8s.io", "v1", "secrets"):                verbsAll,
			NewResourceRequirement("networking.k8s.io", "v1", "services"):               verbsAll,
			NewResourceRequirement("storage.k8s.io", "v1", "pods"):                      verbsGetListWatch,
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
	// These are only required because the role in the Helm chart specifies broad swathes of permissions that
	// don't necessarily exist in the real world.
	RoleOnlyResourceRequirements map[*ResourceRequirement]ResourceVerbs
}

var verbsAll ResourceVerbs = []string{"create", "delete", "deletecollection", "get", "list", "update", "patch", "watch"}
var verbsGetListWatch = []string{"get", "list", "watch"}
var verbsGetCreate = []string{"get", "create"}

// NewResourceRequirement is just a convenience function for creating ResourceRequirements for which the resource type must exist.
func NewResourceRequirement(apiGroup, version, resource string) *ResourceRequirement {
	return &ResourceRequirement{
		Group:    apiGroup,
		Resource: resource,
		Version:  version,
	}
}
