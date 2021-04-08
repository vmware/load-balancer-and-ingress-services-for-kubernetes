package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DockerRegistry docker registry
// swagger:model DockerRegistry
type DockerRegistry struct {

	// Openshift integrated registry config.
	OshiftRegistry *OshiftDockerRegistryMetaData `json:"oshift_registry,omitempty"`

	// Password for docker registry. Authorized 'regular user' password if registry is Openshift integrated registry.
	Password *string `json:"password,omitempty"`

	// Set if docker registry is private. Avi controller will not attempt to push SE image to the registry, unless se_repository_push is set.
	Private *bool `json:"private,omitempty"`

	// Avi ServiceEngine repository name. For private registry, it's registry port/repository, for public registry, it's registry/repository, for openshift registry, it's registry port/namespace/repo.
	Registry *string `json:"registry,omitempty"`

	// Avi Controller will push ServiceEngine image to docker repository. Field deprecated in 18.2.6.
	SeRepositoryPush *bool `json:"se_repository_push,omitempty"`

	// Username for docker registry. Authorized 'regular user' if registry is Openshift integrated registry.
	Username *string `json:"username,omitempty"`
}
