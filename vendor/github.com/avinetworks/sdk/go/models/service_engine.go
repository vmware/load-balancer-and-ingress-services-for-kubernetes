package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServiceEngine service engine
// swagger:model ServiceEngine
type ServiceEngine struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// availability_zone of ServiceEngine.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Placeholder for description of property container_mode of obj type ServiceEngine field type str  type boolean
	ContainerMode *bool `json:"container_mode,omitempty"`

	//  Enum options - CONTAINER_TYPE_BRIDGE, CONTAINER_TYPE_HOST, CONTAINER_TYPE_HOST_DPDK.
	ContainerType *string `json:"container_type,omitempty"`

	// Placeholder for description of property controller_created of obj type ServiceEngine field type str  type boolean
	ControllerCreated *bool `json:"controller_created,omitempty"`

	// controller_ip of ServiceEngine.
	ControllerIP *string `json:"controller_ip,omitempty"`

	// Placeholder for description of property data_vnics of obj type ServiceEngine field type str  type object
	DataVnics []*VNIC `json:"data_vnics,omitempty"`

	// inorder to disable SE set this field appropriately. Enum options - SE_STATE_ENABLED, SE_STATE_DISABLED_FOR_PLACEMENT, SE_STATE_DISABLED, SE_STATE_DISABLED_FORCE.
	EnableState *string `json:"enable_state,omitempty"`

	// flavor of ServiceEngine.
	Flavor *string `json:"flavor,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime.
	HostRef *string `json:"host_ref,omitempty"`

	//  Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN.
	Hypervisor *string `json:"hypervisor,omitempty"`

	// Placeholder for description of property mgmt_vnic of obj type ServiceEngine field type str  type object
	MgmtVnic *VNIC `json:"mgmt_vnic,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Placeholder for description of property resources of obj type ServiceEngine field type str  type object
	Resources *SeResources `json:"resources,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
