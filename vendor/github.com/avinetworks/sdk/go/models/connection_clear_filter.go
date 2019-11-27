package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConnectionClearFilter connection clear filter
// swagger:model ConnectionClearFilter
type ConnectionClearFilter struct {

	// IP address in dotted decimal notation.
	IPAddr *string `json:"ip_addr,omitempty"`

	// Port number.
	Port *int32 `json:"port,omitempty"`
}
