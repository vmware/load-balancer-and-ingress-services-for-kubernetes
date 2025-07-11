// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DockerRegistry docker registry
// swagger:model DockerRegistry
type DockerRegistry struct {

	// Openshift integrated registry config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OshiftRegistry *OshiftDockerRegistryMetaData `json:"oshift_registry,omitempty"`

	// Password for docker registry. Authorized 'regular user' password if registry is Openshift integrated registry. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Password *string `json:"password,omitempty"`

	// Set if docker registry is private. Avi controller will not attempt to push SE image to the registry, unless se_repository_push is set. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Private *bool `json:"private,omitempty"`

	// Avi ServiceEngine repository name. For private registry, it's registry port/repository, for public registry, it's registry/repository, for openshift registry, it's registry port/namespace/repo. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Registry *string `json:"registry,omitempty"`

	// Username for docker registry. Authorized 'regular user' if registry is Openshift integrated registry. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Username *string `json:"username,omitempty"`
}
