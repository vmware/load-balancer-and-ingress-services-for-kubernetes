package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StaticIPRange static Ip range
// swagger:model StaticIpRange
type StaticIPRange struct {

	// IP range. Field introduced in 20.1.3.
	// Required: true
	Range *IPAddrRange `json:"range"`

	// Object type (VIP only, Service Engine only, or both) which can use this IP range. Enum options - STATIC_IPS_FOR_SE, STATIC_IPS_FOR_VIP, STATIC_IPS_FOR_VIP_AND_SE. Field introduced in 20.1.3.
	Type *string `json:"type,omitempty"`
}
