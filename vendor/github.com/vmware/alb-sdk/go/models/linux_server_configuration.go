package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LinuxServerConfiguration linux server configuration
// swagger:model LinuxServerConfiguration
type LinuxServerConfiguration struct {

	// Private docker registry for SE image storage. Field deprecated in 17.1.2.
	DockerRegistrySe *DockerRegistry `json:"docker_registry_se,omitempty"`

	// Placeholder for description of property hosts of obj type LinuxServerConfiguration field type str  type object
	Hosts []*LinuxServerHost `json:"hosts,omitempty"`

	// Flag to notify the SE's in this cloud have an inband management interface, this can be overridden at SE host level by setting host_attr attr_key as SE_INBAND_MGMT with value of true or false.
	SeInbandMgmt *bool `json:"se_inband_mgmt,omitempty"`

	// SE Client Logs disk path for cloud.
	SeLogDiskPath *string `json:"se_log_disk_path,omitempty"`

	// SE Client Log disk size for cloud.
	SeLogDiskSizeGB *int32 `json:"se_log_disk_size_GB,omitempty"`

	// SE System Logs disk path for cloud.
	SeSysDiskPath *string `json:"se_sys_disk_path,omitempty"`

	// SE System Logs disk size for cloud.
	SeSysDiskSizeGB *int32 `json:"se_sys_disk_size_GB,omitempty"`

	// Parameters for SSH to hosts. Field deprecated in 17.1.1.
	SSHAttr *SSHSeDeployment `json:"ssh_attr,omitempty"`

	// Cloud connector user uuid for SSH to hosts. It is a reference to an object of type CloudConnectorUser. Field introduced in 17.1.1.
	SSHUserRef *string `json:"ssh_user_ref,omitempty"`
}
