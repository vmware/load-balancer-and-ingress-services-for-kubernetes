// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SupportedMigrations supported migrations
// swagger:model SupportedMigrations
type SupportedMigrations struct {

	// Minimum accepted API version. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	APIVersion *string `json:"api_version,omitempty"`

	// Minimum space required(in GB) on controller host for this image installation. Field introduced in 18.2.6. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerHostMinFreeDiskSize *int32 `json:"controller_host_min_free_disk_size,omitempty"`

	// Minimum number of cores required for Controller. Field introduced in 18.2.10, 20.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ControllerMinCores *int32 `json:"controller_min_cores,omitempty"`

	// Minimum supported Docker version required for Controller. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ControllerMinDockerVersion *string `json:"controller_min_docker_version,omitempty"`

	// Minimum space required(in GB) on controller for this image installation. Field introduced in 18.2.6. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerMinFreeDiskSize *int32 `json:"controller_min_free_disk_size,omitempty"`

	// Minimum memory required(in GB) for Controller. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ControllerMinMemory *int32 `json:"controller_min_memory,omitempty"`

	// Minimum space required(in GB) for Controller. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ControllerMinTotalDisk *int32 `json:"controller_min_total_disk,omitempty"`

	// Supported active versions for this image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxActiveVersions *int32 `json:"max_active_versions,omitempty"`

	// Minimum supported API version. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinSupportedAPIVersion *string `json:"min_supported_api_version,omitempty"`

	// Minimum space required(in GB) on podman controller host for this image installation. Field introduced in 21.1.4. Unit is GB. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PodmanControllerHostMinFreeDiskSize *int32 `json:"podman_controller_host_min_free_disk_size,omitempty"`

	// Minimum space required(in GB) on podman se host for this image installation. Field introduced in 21.1.4. Unit is GB. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PodmanSeHostMinFreeDiskSize *int32 `json:"podman_se_host_min_free_disk_size,omitempty"`

	// Minimum space required(in GB) on controller for rollback. Field introduced in 18.2.6. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RollbackControllerDiskSpace *int32 `json:"rollback_controller_disk_space,omitempty"`

	// Minimum space required(in GB) on se for rollback. Field introduced in 18.2.6. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RollbackSeDiskSpace *int32 `json:"rollback_se_disk_space,omitempty"`

	// Minimum space required(in GB) on se host for this image installation. Field introduced in 18.2.6. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHostMinFreeDiskSize *int32 `json:"se_host_min_free_disk_size,omitempty"`

	// Minimum  number of cores required for se. Field introduced in 18.2.10, 20.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SeMinCores *int32 `json:"se_min_cores,omitempty"`

	// Minimum space required(in GB) on se for this image installation for non-fips mode(+1 GB for fips mode). Field introduced in 18.2.6. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMinFreeDiskSize *int32 `json:"se_min_free_disk_size,omitempty"`

	// Minimum  memory required(in GB) for se. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SeMinMemory *int32 `json:"se_min_memory,omitempty"`

	// Minimum space required(in GB) for se. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SeMinTotalDisk *int32 `json:"se_min_total_disk,omitempty"`

	// Supported compatible versions for this image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Versions []string `json:"versions,omitempty"`
}
