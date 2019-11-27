package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DockerUCPSetup docker u c p setup
// swagger:model DockerUCPSetup
type DockerUCPSetup struct {

	// cc_id of DockerUCPSetup.
	CcID *string `json:"cc_id,omitempty"`

	// Placeholder for description of property docker_ucp_access of obj type DockerUCPSetup field type str  type boolean
	DockerUcpAccess *bool `json:"docker_ucp_access,omitempty"`

	// failed_hosts of DockerUCPSetup.
	FailedHosts []string `json:"failed_hosts,omitempty"`

	// fleet_endpoint of DockerUCPSetup.
	FleetEndpoint *string `json:"fleet_endpoint,omitempty"`

	// hosts of DockerUCPSetup.
	Hosts []string `json:"hosts,omitempty"`

	// missing_hosts of DockerUCPSetup.
	MissingHosts []string `json:"missing_hosts,omitempty"`

	// new_hosts of DockerUCPSetup.
	NewHosts []string `json:"new_hosts,omitempty"`

	// reason of DockerUCPSetup.
	Reason *string `json:"reason,omitempty"`

	// Placeholder for description of property se_deploy_method_access of obj type DockerUCPSetup field type str  type boolean
	SeDeployMethodAccess *bool `json:"se_deploy_method_access,omitempty"`

	// se_name of DockerUCPSetup.
	SeName *string `json:"se_name,omitempty"`

	// ucp_nodes of DockerUCPSetup.
	UcpNodes []string `json:"ucp_nodes,omitempty"`

	// version of DockerUCPSetup.
	Version *string `json:"version,omitempty"`
}
