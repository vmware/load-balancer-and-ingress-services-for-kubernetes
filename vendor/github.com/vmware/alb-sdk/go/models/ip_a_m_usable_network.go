package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMUsableNetwork ipam usable network
// swagger:model IpamUsableNetwork
type IPAMUsableNetwork struct {

	// Labels as key value pairs, used for selection of IPAM networks. Field introduced in 20.1.3. Maximum of 1 items allowed.
	Labels []*KeyValueTuple `json:"labels,omitempty"`

	// Network. It is a reference to an object of type Network. Field introduced in 20.1.3.
	// Required: true
	NwRef *string `json:"nw_ref"`
}
