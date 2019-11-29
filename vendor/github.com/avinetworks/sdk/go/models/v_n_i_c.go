package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VNIC v n i c
// swagger:model vNIC
type VNIC struct {

	// adapter of vNIC.
	Adapter *string `json:"adapter,omitempty"`

	//  Field introduced in 17.2.7.
	AggregatorChgd *bool `json:"aggregator_chgd,omitempty"`

	// Placeholder for description of property can_se_dp_takeover of obj type vNIC field type str  type boolean
	CanSeDpTakeover *bool `json:"can_se_dp_takeover,omitempty"`

	// Placeholder for description of property connected of obj type vNIC field type str  type boolean
	Connected *bool `json:"connected,omitempty"`

	// Placeholder for description of property del_pending of obj type vNIC field type str  type boolean
	DelPending *bool `json:"del_pending,omitempty"`

	// Placeholder for description of property dhcp_enabled of obj type vNIC field type str  type boolean
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	// Placeholder for description of property enabled of obj type vNIC field type str  type boolean
	Enabled *bool `json:"enabled,omitempty"`

	// if_name of vNIC.
	IfName *string `json:"if_name,omitempty"`

	// Enable IPv6 auto configuration. Field introduced in 18.1.1.
	Ip6AutocfgEnabled *bool `json:"ip6_autocfg_enabled,omitempty"`

	// Placeholder for description of property is_asm of obj type vNIC field type str  type boolean
	IsAsm *bool `json:"is_asm,omitempty"`

	// Placeholder for description of property is_avi_internal_network of obj type vNIC field type str  type boolean
	IsAviInternalNetwork *bool `json:"is_avi_internal_network,omitempty"`

	// Placeholder for description of property is_hsm of obj type vNIC field type str  type boolean
	IsHsm *bool `json:"is_hsm,omitempty"`

	// Placeholder for description of property is_mgmt of obj type vNIC field type str  type boolean
	IsMgmt *bool `json:"is_mgmt,omitempty"`

	// Placeholder for description of property is_portchannel of obj type vNIC field type str  type boolean
	IsPortchannel *bool `json:"is_portchannel,omitempty"`

	// linux_name of vNIC.
	LinuxName *string `json:"linux_name,omitempty"`

	// mac_address of vNIC.
	// Required: true
	MacAddress *string `json:"mac_address"`

	// Placeholder for description of property members of obj type vNIC field type str  type object
	Members []*MemberInterface `json:"members,omitempty"`

	// Number of mtu.
	Mtu *int32 `json:"mtu,omitempty"`

	// network_name of vNIC.
	NetworkName *string `json:"network_name,omitempty"`

	//  It is a reference to an object of type Network.
	NetworkRef *string `json:"network_ref,omitempty"`

	// pci_id of vNIC.
	PciID *string `json:"pci_id,omitempty"`

	// Unique object identifier of port.
	PortUUID *string `json:"port_uuid,omitempty"`

	// Number of vlan_id.
	VlanID *int32 `json:"vlan_id,omitempty"`

	// Placeholder for description of property vlan_interfaces of obj type vNIC field type str  type object
	VlanInterfaces []*VlanInterface `json:"vlan_interfaces,omitempty"`

	// Placeholder for description of property vnic_networks of obj type vNIC field type str  type object
	VnicNetworks []*VNICNetwork `json:"vnic_networks,omitempty"`

	// Number of vrf_id.
	VrfID *int32 `json:"vrf_id,omitempty"`

	//  It is a reference to an object of type VrfContext.
	VrfRef *string `json:"vrf_ref,omitempty"`
}
