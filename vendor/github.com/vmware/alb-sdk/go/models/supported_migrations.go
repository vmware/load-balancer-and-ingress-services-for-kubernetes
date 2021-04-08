package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SupportedMigrations supported migrations
// swagger:model SupportedMigrations
type SupportedMigrations struct {

	// Minimum accepted API version. Field introduced in 18.2.6.
	APIVersion *string `json:"api_version,omitempty"`

	// Minimum space required(in GB) on controller host for this image installation. Field introduced in 18.2.6. Unit is GB.
	ControllerHostMinFreeDiskSize *int32 `json:"controller_host_min_free_disk_size,omitempty"`

	// Minimum number of cores required for Controller. Field introduced in 18.2.10, 20.1.2. Allowed in Basic edition, Essentials edition, Enterprise edition.
	ControllerMinCores *int32 `json:"controller_min_cores,omitempty"`

	// Minimum supported Docker version required for Controller. Field introduced in 21.1.1.
	ControllerMinDockerVersion *string `json:"controller_min_docker_version,omitempty"`

	// Minimum space required(in GB) on controller for this image installation. Field introduced in 18.2.6. Unit is GB.
	ControllerMinFreeDiskSize *int32 `json:"controller_min_free_disk_size,omitempty"`

	// Minimum memory required(in GB) for Controller. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Basic edition, Essentials edition, Enterprise edition.
	ControllerMinMemory *int32 `json:"controller_min_memory,omitempty"`

	// Minimum space required(in GB) for Controller. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Basic edition, Essentials edition, Enterprise edition.
	ControllerMinTotalDisk *int32 `json:"controller_min_total_disk,omitempty"`

	// Supported active versions for this image. Field introduced in 18.2.6.
	MaxActiveVersions *int32 `json:"max_active_versions,omitempty"`

	// Minimum supported API version. Field introduced in 21.1.1.
	MinSupportedAPIVersion *string `json:"min_supported_api_version,omitempty"`

	// Minimum space required(in GB) on controller for rollback. Field introduced in 18.2.6. Unit is GB.
	RollbackControllerDiskSpace *int32 `json:"rollback_controller_disk_space,omitempty"`

	// Minimum space required(in GB) on se for rollback. Field introduced in 18.2.6. Unit is GB.
	RollbackSeDiskSpace *int32 `json:"rollback_se_disk_space,omitempty"`

	// Minimum space required(in GB) on se host for this image installation. Field introduced in 18.2.6. Unit is GB.
	SeHostMinFreeDiskSize *int32 `json:"se_host_min_free_disk_size,omitempty"`

	// Minimum  number of cores required for se. Field introduced in 18.2.10, 20.1.2. Allowed in Basic edition, Essentials edition, Enterprise edition.
	SeMinCores *int32 `json:"se_min_cores,omitempty"`

	// Minimum space required(in GB) on se for this image installation. Field introduced in 18.2.6. Unit is GB.
	SeMinFreeDiskSize *int32 `json:"se_min_free_disk_size,omitempty"`

	// Minimum  memory required(in GB) for se. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Basic edition, Essentials edition, Enterprise edition.
	SeMinMemory *int32 `json:"se_min_memory,omitempty"`

	// Minimum space required(in GB) for se. Field introduced in 18.2.10, 20.1.2. Unit is GB. Allowed in Basic edition, Essentials edition, Enterprise edition.
	SeMinTotalDisk *int32 `json:"se_min_total_disk,omitempty"`

	// Supported compatible versions for this image. Field introduced in 18.2.6.
	Versions []string `json:"versions,omitempty"`
}
