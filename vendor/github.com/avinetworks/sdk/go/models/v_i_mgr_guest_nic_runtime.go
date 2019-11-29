package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrGuestNicRuntime v i mgr guest nic runtime
// swagger:model VIMgrGuestNicRuntime
type VIMgrGuestNicRuntime struct {

	// Placeholder for description of property avi_internal_network of obj type VIMgrGuestNicRuntime field type str  type boolean
	AviInternalNetwork *bool `json:"avi_internal_network,omitempty"`

	// Placeholder for description of property connected of obj type VIMgrGuestNicRuntime field type str  type boolean
	Connected *bool `json:"connected,omitempty"`

	// Placeholder for description of property del_pending of obj type VIMgrGuestNicRuntime field type str  type boolean
	DelPending *bool `json:"del_pending,omitempty"`

	// Placeholder for description of property guest_ip of obj type VIMgrGuestNicRuntime field type str  type object
	GuestIP []*VIMgrIPSubnetRuntime `json:"guest_ip,omitempty"`

	// label of VIMgrGuestNicRuntime.
	Label *string `json:"label,omitempty"`

	// mac_addr of VIMgrGuestNicRuntime.
	// Required: true
	MacAddr *string `json:"mac_addr"`

	// Placeholder for description of property mgmt_vnic of obj type VIMgrGuestNicRuntime field type str  type boolean
	MgmtVnic *bool `json:"mgmt_vnic,omitempty"`

	// network_name of VIMgrGuestNicRuntime.
	NetworkName *string `json:"network_name,omitempty"`

	// Unique object identifier of network.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	// Unique object identifier of os_port.
	OsPortUUID *string `json:"os_port_uuid,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	// Required: true
	Type *string `json:"type"`
}
