package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudVnicChange cloud vnic change
// swagger:model CloudVnicChange
type CloudVnicChange struct {

	// cc_id of CloudVnicChange.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudVnicChange.
	ErrorString *string `json:"error_string,omitempty"`

	// mac_addrs of CloudVnicChange.
	MacAddrs []string `json:"mac_addrs,omitempty"`

	// Unique object identifier of se_vm.
	// Required: true
	SeVMUUID *string `json:"se_vm_uuid"`

	// Placeholder for description of property vnics of obj type CloudVnicChange field type str  type object
	Vnics []*CCVnicInfo `json:"vnics,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
