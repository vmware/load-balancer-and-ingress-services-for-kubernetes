package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrNWRuntime v i mgr n w runtime
// swagger:model VIMgrNWRuntime
type VIMgrNWRuntime struct {

	// Placeholder for description of property MgmtNW of obj type VIMgrNWRuntime field type str  type boolean
	MgmtNW *bool `json:"MgmtNW,omitempty"`

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// apic_vrf_context of VIMgrNWRuntime.
	ApicVrfContext *string `json:"apic_vrf_context,omitempty"`

	// Placeholder for description of property auto_expand of obj type VIMgrNWRuntime field type str  type boolean
	AutoExpand *bool `json:"auto_expand,omitempty"`

	// availability_zone of VIMgrNWRuntime.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Unique object identifier of datacenter.
	DatacenterUUID *string `json:"datacenter_uuid,omitempty"`

	// Placeholder for description of property dvs of obj type VIMgrNWRuntime field type str  type boolean
	Dvs *bool `json:"dvs,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime.
	HostRefs []string `json:"host_refs,omitempty"`

	// Placeholder for description of property interested_nw of obj type VIMgrNWRuntime field type str  type boolean
	InterestedNw *bool `json:"interested_nw,omitempty"`

	// Placeholder for description of property ip_subnet of obj type VIMgrNWRuntime field type str  type object
	IPSubnet []*VIMgrIPSubnetRuntime `json:"ip_subnet,omitempty"`

	// managed_object_id of VIMgrNWRuntime.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Number of num_ports.
	NumPorts *int32 `json:"num_ports,omitempty"`

	// switch_name of VIMgrNWRuntime.
	SwitchName *string `json:"switch_name,omitempty"`

	// tenant_name of VIMgrNWRuntime.
	TenantName *string `json:"tenant_name,omitempty"`

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

	// Number of vlan.
	Vlan *int32 `json:"vlan,omitempty"`

	// Placeholder for description of property vlan_range of obj type VIMgrNWRuntime field type str  type object
	VlanRange []*VlanRange `json:"vlan_range,omitempty"`

	//  It is a reference to an object of type VIMgrVMRuntime.
	VMRefs []string `json:"vm_refs,omitempty"`

	//  It is a reference to an object of type VrfContext.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`
}
