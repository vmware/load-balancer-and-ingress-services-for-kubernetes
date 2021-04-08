package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NatAddrInfo nat addr info
// swagger:model NatAddrInfo
type NatAddrInfo struct {

	// Nat IP address. Field introduced in 18.2.3.
	NatIP *IPAddr `json:"nat_ip,omitempty"`

	// Nat IP address range. Field introduced in 18.2.3.
	NatIPRange *IPAddrRange `json:"nat_ip_range,omitempty"`
}
