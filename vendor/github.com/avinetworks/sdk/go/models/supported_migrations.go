package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SupportedMigrations supported migrations
// swagger:model SupportedMigrations
type SupportedMigrations struct {

	// Api version of the image. Field introduced in 18.2.6.
	APIVersion *string `json:"api_version,omitempty"`

	// Minimum space required(in GB) on controller host for this image installation. Field introduced in 18.2.6.
	ControllerHostMinFreeDiskSize *int32 `json:"controller_host_min_free_disk_size,omitempty"`

	// Minimum space required(in GB) on controller for this image installation. Field introduced in 18.2.6.
	ControllerMinFreeDiskSize *int32 `json:"controller_min_free_disk_size,omitempty"`

	// Supported active versions for this image. Field introduced in 18.2.6.
	MaxActiveVersions *int32 `json:"max_active_versions,omitempty"`

	// Minimum space required(in GB) on controller for rollback. Field introduced in 18.2.6.
	RollbackControllerDiskSpace *int32 `json:"rollback_controller_disk_space,omitempty"`

	// Minimum space required(in GB) on se for rollback. Field introduced in 18.2.6.
	RollbackSeDiskSpace *int32 `json:"rollback_se_disk_space,omitempty"`

	// Minimum space required(in GB) on se host for this image installation. Field introduced in 18.2.6.
	SeHostMinFreeDiskSize *int32 `json:"se_host_min_free_disk_size,omitempty"`

	// Minimum space required(in GB) on se for this image installation. Field introduced in 18.2.6.
	SeMinFreeDiskSize *int32 `json:"se_min_free_disk_size,omitempty"`

	// Supported compatible versions for this image. Field introduced in 18.2.6.
	Versions []string `json:"versions,omitempty"`
}
