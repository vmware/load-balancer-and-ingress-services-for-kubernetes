// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerSizingCloudLimits controller sizing cloud limits
// swagger:model ControllerSizingCloudLimits
type ControllerSizingCloudLimits struct {

	// Maximum number of clouds of a given type. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumClouds *int32 `json:"num_clouds,omitempty"`

	// Cloud type for the limit. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`
}
