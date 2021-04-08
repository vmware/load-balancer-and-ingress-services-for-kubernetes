package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeBgpPeerDownDetails se bgp peer down details
// swagger:model SeBgpPeerDownDetails
type SeBgpPeerDownDetails struct {

	// Message specific to the down condition. Field introduced in 20.1.1.
	Message *string `json:"message,omitempty"`

	// IP address of BGP peer. Field introduced in 20.1.1.
	// Required: true
	PeerIP *string `json:"peer_ip"`

	// Name of Virtual Routing Context in which BGP is configured. Field introduced in 20.1.1.
	VrfName *string `json:"vrf_name,omitempty"`
}
