package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugIPAddr debug Ip addr
// swagger:model DebugIpAddr
type DebugIPAddr struct {

	// Placeholder for description of property addrs of obj type DebugIpAddr field type str  type object
	Addrs []*IPAddr `json:"addrs,omitempty"`

	// Placeholder for description of property prefixes of obj type DebugIpAddr field type str  type object
	Prefixes []*IPAddrPrefix `json:"prefixes,omitempty"`

	// Placeholder for description of property ranges of obj type DebugIpAddr field type str  type object
	Ranges []*IPAddrRange `json:"ranges,omitempty"`
}
