package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtHosts nsxt hosts
// swagger:model NsxtHosts
type NsxtHosts struct {

	// List of transport nodes. Field introduced in 20.1.1.
	HostIds []string `json:"host_ids,omitempty"`

	// Include or Exclude. Field introduced in 20.1.1.
	Include *bool `json:"include,omitempty"`
}
