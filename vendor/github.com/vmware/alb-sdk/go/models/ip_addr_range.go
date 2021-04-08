package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAddrRange Ip addr range
// swagger:model IpAddrRange
type IPAddrRange struct {

	// Starting IP address of the range.
	// Required: true
	Begin *IPAddr `json:"begin"`

	// Ending IP address of the range.
	// Required: true
	End *IPAddr `json:"end"`
}
