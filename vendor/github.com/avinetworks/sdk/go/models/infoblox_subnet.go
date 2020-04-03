package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// InfobloxSubnet infoblox subnet
// swagger:model InfobloxSubnet
type InfobloxSubnet struct {

	// IPv4 subnet to use for Infoblox allocation. Field introduced in 18.2.8.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// IPv6 subnet to use for Infoblox allocation. Field introduced in 18.2.8.
	Subnet6 *IPAddrPrefix `json:"subnet6,omitempty"`
}
