package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIFaultInjection v i fault injection
// swagger:model VIFaultInjection
type VIFaultInjection struct {

	//  Enum options - INITIAL_VALUE. CREATE_SE. MODIFY_VNIC. VM_MONITOR. RESOURCE_MONITOR. PERF_MONITOR. SET_MGMT_IP. MODIFY_MGMT_IP. SIM_VM_BULK_NOTIF. RESYNC_ERROR. SIMULATE_OVA_ERR. VCENTER_NO_OBJECTS. CREATE_VM_RUNTIME_ERR. VERSION_NULL_ERR. DISC_PGNAME_ERR. DISC_DCDETAILS_ERR. DISC_DC_ERR. DISC_HOST_ERR. DISC_CLUSTER_ERR. DISC_PG_ERR...
	// Required: true
	API *string `json:"api"`

	//  Field introduced in 17.1.3.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// Number of count.
	// Required: true
	Count *int32 `json:"count"`

	//  Field introduced in 17.1.3.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	//  Enum options - SEVM_SUCCESS. SEVM_CREATE_FAIL_NO_HW_INFO. SEVM_CREATE_FAIL_DUPLICATE_NAME. SEVM_CREATE_FAIL_NO_MGMT_NW. SEVM_CREATE_FAIL_NO_CPU. SEVM_CREATE_FAIL_NO_MEM. SEVM_CREATE_FAIL_NO_LEASE. SEVM_CREATE_FAIL_OVF_ERROR. SEVM_CREATE_NO_HOST_VM_NETWORK. SEVM_CREATE_FAIL_NO_PROGRESS. SEVM_CREATE_FAIL_ABORTED. SEVM_CREATE_FAILURE. SEVM_CREATE_FAIL_POWER_ON. SEVM_VNIC_NO_VM. SEVM_VNIC_MAC_ADDR_ERROR. SEVM_VNIC_FAILURE. SEVM_VNIC_NO_PG_PORTS. SEVM_DELETE_FAILURE. SEVM_CREATE_LIMIT_REACHED. SEVM_SET_MGMT_IP_FAILED...
	Status *string `json:"status,omitempty"`
}
