package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ContainerCloudSetup container cloud setup
// swagger:model ContainerCloudSetup
type ContainerCloudSetup struct {

	// cc_id of ContainerCloudSetup.
	CcID *string `json:"cc_id,omitempty"`

	// Placeholder for description of property cloud_access of obj type ContainerCloudSetup field type str  type boolean
	CloudAccess *bool `json:"cloud_access,omitempty"`

	// failed_hosts of ContainerCloudSetup.
	FailedHosts []string `json:"failed_hosts,omitempty"`

	// fleet_endpoint of ContainerCloudSetup.
	FleetEndpoint *string `json:"fleet_endpoint,omitempty"`

	// hosts of ContainerCloudSetup.
	Hosts []string `json:"hosts,omitempty"`

	// master_nodes of ContainerCloudSetup.
	MasterNodes []string `json:"master_nodes,omitempty"`

	// missing_hosts of ContainerCloudSetup.
	MissingHosts []string `json:"missing_hosts,omitempty"`

	// new_hosts of ContainerCloudSetup.
	NewHosts []string `json:"new_hosts,omitempty"`

	// reason of ContainerCloudSetup.
	Reason *string `json:"reason,omitempty"`

	// Placeholder for description of property se_deploy_method_access of obj type ContainerCloudSetup field type str  type boolean
	SeDeployMethodAccess *bool `json:"se_deploy_method_access,omitempty"`

	// se_name of ContainerCloudSetup.
	SeName *string `json:"se_name,omitempty"`

	// version of ContainerCloudSetup.
	Version *string `json:"version,omitempty"`
}
