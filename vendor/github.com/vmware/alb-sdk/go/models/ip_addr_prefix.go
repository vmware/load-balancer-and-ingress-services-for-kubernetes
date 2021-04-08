package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAddrPrefix Ip addr prefix
// swagger:model IpAddrPrefix
type IPAddrPrefix struct {

	// Placeholder for description of property ip_addr of obj type IpAddrPrefix field type str  type object
	// Required: true
	IPAddr *IPAddr `json:"ip_addr"`

	// Number of mask.
	// Required: true
	Mask *int32 `json:"mask"`
}
