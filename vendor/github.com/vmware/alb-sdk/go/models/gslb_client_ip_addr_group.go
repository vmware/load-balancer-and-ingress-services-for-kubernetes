package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbClientIPAddrGroup gslb client Ip addr group
// swagger:model GslbClientIpAddrGroup
type GslbClientIPAddrGroup struct {

	// Configure IP address(es). Field introduced in 17.1.2.
	Addrs []*IPAddr `json:"addrs,omitempty"`

	// Configure IP address prefix(es). Field introduced in 17.1.2.
	Prefixes []*IPAddrPrefix `json:"prefixes,omitempty"`

	// Configure IP address range(s). Field introduced in 17.1.2.
	Ranges []*IPAddrRange `json:"ranges,omitempty"`

	// Specify whether this client IP address range is public or private. Enum options - GSLB_IP_PUBLIC, GSLB_IP_PRIVATE. Field introduced in 17.1.2.
	// Required: true
	Type *string `json:"type"`
}
