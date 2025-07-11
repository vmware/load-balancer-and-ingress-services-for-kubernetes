// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MarathonSeDeployment marathon se deployment
// swagger:model MarathonSeDeployment
type MarathonSeDeployment struct {

	// Docker image to be used for Avi SE installation e.g. fedora, ubuntu. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerImage *string `json:"docker_image,omitempty"`

	// Host OS distribution e.g. COREOS, UBUNTU, REDHAT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostOs *string `json:"host_os,omitempty"`

	// Accepted resource roles for SEs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceRoles []string `json:"resource_roles,omitempty"`

	// URIs to be resolved for starting the application. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Uris []string `json:"uris,omitempty"`
}
