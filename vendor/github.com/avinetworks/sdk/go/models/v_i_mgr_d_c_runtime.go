package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrDCRuntime v i mgr d c runtime
// swagger:model VIMgrDCRuntime
type VIMgrDCRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	//  It is a reference to an object of type VIMgrClusterRuntime.
	ClusterRefs []string `json:"cluster_refs,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime.
	HostRefs []string `json:"host_refs,omitempty"`

	// Placeholder for description of property interested_hosts of obj type VIMgrDCRuntime field type str  type object
	InterestedHosts []*VIMgrInterestedEntity `json:"interested_hosts,omitempty"`

	// Placeholder for description of property interested_nws of obj type VIMgrDCRuntime field type str  type object
	InterestedNws []*VIMgrInterestedEntity `json:"interested_nws,omitempty"`

	// Placeholder for description of property interested_vms of obj type VIMgrDCRuntime field type str  type object
	InterestedVms []*VIMgrInterestedEntity `json:"interested_vms,omitempty"`

	// Number of inventory_state.
	InventoryState *int32 `json:"inventory_state,omitempty"`

	// managed_object_id of VIMgrDCRuntime.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type VIMgrNWRuntime.
	NwRefs []string `json:"nw_refs,omitempty"`

	// Number of pending_vcenter_reqs.
	PendingVcenterReqs *int32 `json:"pending_vcenter_reqs,omitempty"`

	//  It is a reference to an object of type VIMgrSEVMRuntime.
	SevmRefs []string `json:"sevm_refs,omitempty"`

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

	// Unique object identifier of vcenter.
	VcenterUUID *string `json:"vcenter_uuid,omitempty"`

	//  It is a reference to an object of type VIMgrVMRuntime.
	VMRefs []string `json:"vm_refs,omitempty"`
}
