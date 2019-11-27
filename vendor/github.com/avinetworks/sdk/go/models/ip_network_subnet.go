package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPNetworkSubnet IP network subnet
// swagger:model IPNetworkSubnet
type IPNetworkSubnet struct {

	// Network for VirtualService IP allocation with Vantage as the IPAM provider. Network should be created before this is configured. It is a reference to an object of type Network.
	NetworkRef *string `json:"network_ref,omitempty"`

	// Subnet for VirtualService IP allocation with Vantage or Infoblox as the IPAM provider. Only one of subnet or subnet_uuid configuration is allowed.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// Subnet for VirtualService IPv6 allocation with Vantage or Infoblox as the IPAM provider. Only one of subnet or subnet_uuid configuration is allowed. Field introduced in 18.1.1.
	Subnet6 *IPAddrPrefix `json:"subnet6,omitempty"`

	// Subnet UUID or Name or Prefix for VirtualService IPv6 allocation with AWS or OpenStack as the IPAM provider. Only one of subnet or subnet_uuid configuration is allowed. Field introduced in 18.1.1.
	Subnet6UUID *string `json:"subnet6_uuid,omitempty"`

	// Subnet UUID or Name or Prefix for VirtualService IP allocation with AWS or OpenStack as the IPAM provider. Only one of subnet or subnet_uuid configuration is allowed.
	SubnetUUID *string `json:"subnet_uuid,omitempty"`
}
