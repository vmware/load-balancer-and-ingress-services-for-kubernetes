package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MesosSetup mesos setup
// swagger:model MesosSetup
type MesosSetup struct {

	// cc_id of MesosSetup.
	CcID *string `json:"cc_id,omitempty"`

	// failed_hosts of MesosSetup.
	FailedHosts []string `json:"failed_hosts,omitempty"`

	// fleet_endpoint of MesosSetup.
	FleetEndpoint *string `json:"fleet_endpoint,omitempty"`

	// hosts of MesosSetup.
	Hosts []string `json:"hosts,omitempty"`

	// Placeholder for description of property mesos_access of obj type MesosSetup field type str  type boolean
	MesosAccess *bool `json:"mesos_access,omitempty"`

	// mesos_url of MesosSetup.
	MesosURL *string `json:"mesos_url,omitempty"`

	// missing_hosts of MesosSetup.
	MissingHosts []string `json:"missing_hosts,omitempty"`

	// new_hosts of MesosSetup.
	NewHosts []string `json:"new_hosts,omitempty"`

	// reason of MesosSetup.
	Reason *string `json:"reason,omitempty"`

	// Placeholder for description of property se_deploy_method_access of obj type MesosSetup field type str  type boolean
	SeDeployMethodAccess *bool `json:"se_deploy_method_access,omitempty"`

	// se_name of MesosSetup.
	SeName *string `json:"se_name,omitempty"`

	// version of MesosSetup.
	Version *string `json:"version,omitempty"`
}
