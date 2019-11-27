package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudHealth cloud health
// swagger:model CloudHealth
type CloudHealth struct {

	// cc_id of CloudHealth.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudHealth.
	ErrorString *string `json:"error_string,omitempty"`

	// first_fail of CloudHealth.
	FirstFail *string `json:"first_fail,omitempty"`

	// last_fail of CloudHealth.
	LastFail *string `json:"last_fail,omitempty"`

	// last_ok of CloudHealth.
	LastOk *string `json:"last_ok,omitempty"`

	// Number of num_fails.
	NumFails *int32 `json:"num_fails,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
