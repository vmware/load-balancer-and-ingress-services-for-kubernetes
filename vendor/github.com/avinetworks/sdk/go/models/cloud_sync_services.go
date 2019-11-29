package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudSyncServices cloud sync services
// swagger:model CloudSyncServices
type CloudSyncServices struct {

	// cc_id of CloudSyncServices.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudSyncServices.
	ErrorString *string `json:"error_string,omitempty"`

	// Unique object identifier of se_vm.
	SeVMUUID *string `json:"se_vm_uuid,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	Vtype *string `json:"vtype,omitempty"`
}
