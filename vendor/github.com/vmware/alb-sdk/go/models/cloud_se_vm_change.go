package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudSeVMChange cloud se Vm change
// swagger:model CloudSeVmChange
type CloudSeVMChange struct {

	// cc_id of CloudSeVmChange.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudSeVmChange.
	ErrorString *string `json:"error_string,omitempty"`

	// Unique object identifier of se_vm.
	SeVMUUID *string `json:"se_vm_uuid,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT.
	Vtype *string `json:"vtype,omitempty"`
}
