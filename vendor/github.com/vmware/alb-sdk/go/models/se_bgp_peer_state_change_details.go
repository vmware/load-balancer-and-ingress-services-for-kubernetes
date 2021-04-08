package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeBgpPeerStateChangeDetails se bgp peer state change details
// swagger:model SeBgpPeerStateChangeDetails
type SeBgpPeerStateChangeDetails struct {

	// IP address of BGP peer. Field introduced in 17.2.1.
	// Required: true
	PeerIP *string `json:"peer_ip"`

	// BGP peer state. Field introduced in 17.2.1.
	// Required: true
	PeerState *string `json:"peer_state"`

	// Name of Virtual Routing Context in which BGP is configured. Field introduced in 17.2.1.
	VrfName *string `json:"vrf_name,omitempty"`
}
