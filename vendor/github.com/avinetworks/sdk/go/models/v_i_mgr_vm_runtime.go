package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrVMRuntime v i mgr VM runtime
// swagger:model VIMgrVMRuntime
type VIMgrVMRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// availability_zone of VIMgrVMRuntime.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// connection_state of VIMgrVMRuntime.
	ConnectionState *string `json:"connection_state,omitempty"`

	// Unique object identifier of controller_cluster.
	ControllerClusterUUID *string `json:"controller_cluster_uuid,omitempty"`

	// controller_ip_addr of VIMgrVMRuntime.
	ControllerIPAddr *string `json:"controller_ip_addr,omitempty"`

	// Placeholder for description of property controller_vm of obj type VIMgrVMRuntime field type str  type boolean
	ControllerVM *bool `json:"controller_vm,omitempty"`

	// Number of cpu_reservation.
	CPUReservation *int64 `json:"cpu_reservation,omitempty"`

	// Number of cpu_shares.
	CPUShares *int32 `json:"cpu_shares,omitempty"`

	// Placeholder for description of property creation_in_progress of obj type VIMgrVMRuntime field type str  type boolean
	CreationInProgress *bool `json:"creation_in_progress,omitempty"`

	// Placeholder for description of property guest_nic of obj type VIMgrVMRuntime field type str  type object
	GuestNic []*VIMgrGuestNicRuntime `json:"guest_nic,omitempty"`

	// host of VIMgrVMRuntime.
	Host *string `json:"host,omitempty"`

	// Number of init_vnics.
	InitVnics *int32 `json:"init_vnics,omitempty"`

	// managed_object_id of VIMgrVMRuntime.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	// Number of mem_shares.
	MemShares *int32 `json:"mem_shares,omitempty"`

	// Number of memory.
	Memory *int64 `json:"memory,omitempty"`

	// Number of memory_reservation.
	MemoryReservation *int64 `json:"memory_reservation,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Number of num_cpu.
	NumCPU *int32 `json:"num_cpu,omitempty"`

	//  Field introduced in 17.1.1,17.1.3.
	OvfAvisetypeField *string `json:"ovf_avisetype_field,omitempty"`

	// powerstate of VIMgrVMRuntime.
	Powerstate *string `json:"powerstate,omitempty"`

	// Number of se_ver.
	SeVer *int32 `json:"se_ver,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Unique object identifier of vcenter_datacenter.
	VcenterDatacenterUUID *string `json:"vcenter_datacenter_uuid,omitempty"`

	// vcenter_rm_cookie of VIMgrVMRuntime.
	VcenterRmCookie *string `json:"vcenter_rm_cookie,omitempty"`

	//  Enum options - VIMGR_SE_NETWORK_ADMIN, VIMGR_SE_UNIFIED_ADMIN.
	VcenterSeType *string `json:"vcenter_se_type,omitempty"`

	// Placeholder for description of property vcenter_template_vm of obj type VIMgrVMRuntime field type str  type boolean
	VcenterTemplateVM *bool `json:"vcenter_template_vm,omitempty"`

	// vcenter_vAppName of VIMgrVMRuntime.
	VcenterVAppName *string `json:"vcenter_vAppName,omitempty"`

	// vcenter_vAppVendor of VIMgrVMRuntime.
	VcenterVAppVendor *string `json:"vcenter_vAppVendor,omitempty"`

	//  Enum options - VMTYPE_SE_VM, VMTYPE_POOL_SRVR.
	VcenterVMType *string `json:"vcenter_vm_type,omitempty"`

	// Placeholder for description of property vcenter_vnic_discovered of obj type VIMgrVMRuntime field type str  type boolean
	VcenterVnicDiscovered *bool `json:"vcenter_vnic_discovered,omitempty"`

	// Number of vm_lb_weight.
	VMLbWeight *int32 `json:"vm_lb_weight,omitempty"`
}
