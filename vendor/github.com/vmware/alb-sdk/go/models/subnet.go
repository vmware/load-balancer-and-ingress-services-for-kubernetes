package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Subnet subnet
// swagger:model Subnet
type Subnet struct {

	// Specify an IP subnet prefix for this Network.
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// Static IP ranges for this subnet. Field introduced in 20.1.3.
	StaticIPRanges []*StaticIPRange `json:"static_ip_ranges,omitempty"`

	// Use static_ip_ranges. Field deprecated in 20.1.3.
	StaticIps []*IPAddr `json:"static_ips,omitempty"`

	// Use static_ip_ranges. Field deprecated in 20.1.3.
	StaticRanges []*IPAddrRange `json:"static_ranges,omitempty"`
}
