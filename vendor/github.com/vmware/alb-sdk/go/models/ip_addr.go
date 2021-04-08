package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAddr Ip addr
// swagger:model IpAddr
type IPAddr struct {

	// IP address.
	// Required: true
	Addr *string `json:"addr"`

	//  Enum options - V4, DNS, V6.
	// Required: true
	Type *string `json:"type"`
}
