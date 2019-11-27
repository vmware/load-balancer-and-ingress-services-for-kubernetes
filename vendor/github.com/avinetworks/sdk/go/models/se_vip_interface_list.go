package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeVipInterfaceList se vip interface list
// swagger:model SeVipInterfaceList
type SeVipInterfaceList struct {

	// Placeholder for description of property is_portchannel of obj type SeVipInterfaceList field type str  type boolean
	IsPortchannel *bool `json:"is_portchannel,omitempty"`

	// Placeholder for description of property vip_intf_ip of obj type SeVipInterfaceList field type str  type object
	VipIntfIP *IPAddr `json:"vip_intf_ip,omitempty"`

	// Placeholder for description of property vip_intf_ip6 of obj type SeVipInterfaceList field type str  type object
	VipIntfIp6 *IPAddr `json:"vip_intf_ip6,omitempty"`

	// vip_intf_mac of SeVipInterfaceList.
	// Required: true
	VipIntfMac *string `json:"vip_intf_mac"`

	// Number of vlan_id.
	VlanID *int32 `json:"vlan_id,omitempty"`
}
