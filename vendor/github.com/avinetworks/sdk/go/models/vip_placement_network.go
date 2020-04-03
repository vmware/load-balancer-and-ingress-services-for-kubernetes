package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VipPlacementNetwork vip placement network
// swagger:model VipPlacementNetwork
type VipPlacementNetwork struct {

	// Network to use for vip placement. It is a reference to an object of type Network. Field introduced in 18.2.5.
	NetworkRef *string `json:"network_ref,omitempty"`

	// IPv4 Subnet to use for vip placement. Field introduced in 18.2.5.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// IPv6 subnet to use for vip placement. Field introduced in 18.2.5.
	Subnet6 *IPAddrPrefix `json:"subnet6,omitempty"`
}
