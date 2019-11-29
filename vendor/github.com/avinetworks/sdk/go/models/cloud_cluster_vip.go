package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudClusterVip cloud cluster vip
// swagger:model CloudClusterVip
type CloudClusterVip struct {

	// cc_id of CloudClusterVip.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudClusterVip.
	ErrorString *string `json:"error_string,omitempty"`

	// Placeholder for description of property ip of obj type CloudClusterVip field type str  type object
	// Required: true
	IP *IPAddr `json:"ip"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
