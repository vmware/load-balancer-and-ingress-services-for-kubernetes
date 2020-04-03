package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrSEVMRuntime v i mgr s e VM runtime
// swagger:model VIMgrSEVMRuntime
type VIMgrSEVMRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// availability_zone of VIMgrSEVMRuntime.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	//  Field introduced in 17.2.1.
	AzureInfo *AzureInfo `json:"azure_info,omitempty"`

	// cloud_name of VIMgrSEVMRuntime.
	CloudName *string `json:"cloud_name,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// connection_state of VIMgrSEVMRuntime.
	ConnectionState *string `json:"connection_state,omitempty"`

	// Unique object identifier of controller_cluster.
	ControllerClusterUUID *string `json:"controller_cluster_uuid,omitempty"`

	// controller_ip_addr of VIMgrSEVMRuntime.
	ControllerIPAddr *string `json:"controller_ip_addr,omitempty"`

	// Service Engine Cookie set by the resource manager. Field introduced in 18.2.2.
	Cookie *string `json:"cookie,omitempty"`

	// Placeholder for description of property creation_in_progress of obj type VIMgrSEVMRuntime field type str  type boolean
	CreationInProgress *bool `json:"creation_in_progress,omitempty"`

	// Placeholder for description of property deletion_in_progress of obj type VIMgrSEVMRuntime field type str  type boolean
	DeletionInProgress *bool `json:"deletion_in_progress,omitempty"`

	// discovery_response of VIMgrSEVMRuntime.
	DiscoveryResponse *string `json:"discovery_response,omitempty"`

	// Number of discovery_status.
	DiscoveryStatus *int32 `json:"discovery_status,omitempty"`

	// Disk space in GB for each service engine VM. Field introduced in 18.2.2.
	DiskGb *int32 `json:"disk_gb,omitempty"`

	// flavor of VIMgrSEVMRuntime.
	Flavor *string `json:"flavor,omitempty"`

	// Placeholder for description of property guest_nic of obj type VIMgrSEVMRuntime field type str  type object
	GuestNic []*VIMgrGuestNicRuntime `json:"guest_nic,omitempty"`

	// host of VIMgrSEVMRuntime.
	Host *string `json:"host,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime.
	HostRef *string `json:"host_ref,omitempty"`

	// hostid of VIMgrSEVMRuntime.
	Hostid *string `json:"hostid,omitempty"`

	//  Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN.
	Hypervisor *string `json:"hypervisor,omitempty"`

	// Number of init_vnics.
	InitVnics *int32 `json:"init_vnics,omitempty"`

	// Number of last_discovery.
	LastDiscovery *int32 `json:"last_discovery,omitempty"`

	// managed_object_id of VIMgrSEVMRuntime.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	// Memory in MB for each service engine VM. Field introduced in 18.2.2.
	MemoryMb *int32 `json:"memory_mb,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// powerstate of VIMgrSEVMRuntime.
	Powerstate *string `json:"powerstate,omitempty"`

	// Unique object identifier of security_group.
	SecurityGroupUUID *string `json:"security_group_uuid,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup.
	SegroupRef *string `json:"segroup_ref,omitempty"`

	// Unique object identifier of server_group.
	ServerGroupUUID *string `json:"server_group_uuid,omitempty"`

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

	// vcenter_rm_cookie of VIMgrSEVMRuntime.
	VcenterRmCookie *string `json:"vcenter_rm_cookie,omitempty"`

	//  Enum options - VIMGR_SE_NETWORK_ADMIN, VIMGR_SE_UNIFIED_ADMIN.
	VcenterSeType *string `json:"vcenter_se_type,omitempty"`

	// Placeholder for description of property vcenter_template_vm of obj type VIMgrSEVMRuntime field type str  type boolean
	VcenterTemplateVM *bool `json:"vcenter_template_vm,omitempty"`

	// vcenter_vAppName of VIMgrSEVMRuntime.
	VcenterVAppName *string `json:"vcenter_vAppName,omitempty"`

	// vcenter_vAppVendor of VIMgrSEVMRuntime.
	VcenterVAppVendor *string `json:"vcenter_vAppVendor,omitempty"`

	//  Enum options - VMTYPE_SE_VM, VMTYPE_POOL_SRVR.
	VcenterVMType *string `json:"vcenter_vm_type,omitempty"`

	// Count of vcpus for each service engine VM. Field introduced in 18.2.2.
	Vcpus *int32 `json:"vcpus,omitempty"`
}
