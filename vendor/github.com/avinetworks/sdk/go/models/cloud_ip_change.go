package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudIPChange cloud Ip change
// swagger:model CloudIpChange
type CloudIPChange struct {

	// cc_id of CloudIpChange.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudIpChange.
	ErrorString *string `json:"error_string,omitempty"`

	// Placeholder for description of property ip of obj type CloudIpChange field type str  type object
	// Required: true
	IP *IPAddr `json:"ip"`

	//  Field introduced in 18.1.1.
	Ip6 *IPAddr `json:"ip6,omitempty"`

	//  Field introduced in 18.1.1.
	Ip6Mask *int32 `json:"ip6_mask,omitempty"`

	//  Field introduced in 17.1.1.
	IPMask *int32 `json:"ip_mask,omitempty"`

	// mac_addr of CloudIpChange.
	MacAddr *string `json:"mac_addr,omitempty"`

	// Unique object identifier of port.
	PortUUID *string `json:"port_uuid,omitempty"`

	// Unique object identifier of se_vm.
	SeVMUUID *string `json:"se_vm_uuid,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
