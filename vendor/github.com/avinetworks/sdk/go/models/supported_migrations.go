package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SupportedMigrations supported migrations
// swagger:model SupportedMigrations
type SupportedMigrations struct {

	// Api version of the image. Field introduced in 18.2.6.
	APIVersion *string `json:"api_version,omitempty"`

	// Minimum space required(in GB) on controller host for this image installation. Field introduced in 18.2.6. Unit is GB.
	ControllerHostMinFreeDiskSize *int32 `json:"controller_host_min_free_disk_size,omitempty"`

	// Minimum number of cores required for Controller. Field introduced in 18.2.10, 20.1.2.
	ControllerMinCores *int32 `json:"controller_min_cores,omitempty"`

	// Minimum space required(in GB) on controller for this image installation. Field introduced in 18.2.6. Unit is GB.
	ControllerMinFreeDiskSize *int32 `json:"controller_min_free_disk_size,omitempty"`

	// Minimum memory required(in GB) for Controller. Field introduced in 18.2.10, 20.1.2. Unit is GB.
	ControllerMinMemory *int32 `json:"controller_min_memory,omitempty"`

	// Minimum space required(in GB) for Controller. Field introduced in 18.2.10, 20.1.2. Unit is GB.
	ControllerMinTotalDisk *int32 `json:"controller_min_total_disk,omitempty"`

	// Supported active versions for this image. Field introduced in 18.2.6.
	MaxActiveVersions *int32 `json:"max_active_versions,omitempty"`

	// Minimum space required(in GB) on controller for rollback. Field introduced in 18.2.6. Unit is GB.
	RollbackControllerDiskSpace *int32 `json:"rollback_controller_disk_space,omitempty"`

	// Minimum space required(in GB) on se for rollback. Field introduced in 18.2.6. Unit is GB.
	RollbackSeDiskSpace *int32 `json:"rollback_se_disk_space,omitempty"`

	// Minimum space required(in GB) on se host for this image installation. Field introduced in 18.2.6. Unit is GB.
	SeHostMinFreeDiskSize *int32 `json:"se_host_min_free_disk_size,omitempty"`

	// Minimum  number of cores required for se. Field introduced in 18.2.10, 20.1.2.
	SeMinCores *int32 `json:"se_min_cores,omitempty"`

	// Minimum space required(in GB) on se for this image installation. Field introduced in 18.2.6. Unit is GB.
	SeMinFreeDiskSize *int32 `json:"se_min_free_disk_size,omitempty"`

	// Minimum  memory required(in GB) for se. Field introduced in 18.2.10, 20.1.2. Unit is GB.
	SeMinMemory *int32 `json:"se_min_memory,omitempty"`

	// Minimum space required(in GB) for se. Field introduced in 18.2.10, 20.1.2. Unit is GB.
	SeMinTotalDisk *int32 `json:"se_min_total_disk,omitempty"`

	// Supported compatible versions for this image. Field introduced in 18.2.6.
	Versions []string `json:"versions,omitempty"`
}
