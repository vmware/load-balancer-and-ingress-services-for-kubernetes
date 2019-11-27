package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Subnet subnet
// swagger:model Subnet
type Subnet struct {

	// Specify an IP subnet prefix for this Network.
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// Specify a pool of IP addresses for use in Service Engines.
	StaticIps []*IPAddr `json:"static_ips,omitempty"`

	// Placeholder for description of property static_ranges of obj type Subnet field type str  type object
	StaticRanges []*IPAddrRange `json:"static_ranges,omitempty"`
}
