package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BgpProfile bgp profile
// swagger:model BgpProfile
type BgpProfile struct {

	// Community *string either in aa nn format where aa, nn is within [1,65535] or local-AS|no-advertise|no-export|internet. Field introduced in 17.1.2. Maximum of 16 items allowed.
	Community []string `json:"community,omitempty"`

	// Hold time for Peers. Allowed values are 3-7200.
	HoldTime *int32 `json:"hold_time,omitempty"`

	// BGP peer type.
	// Required: true
	Ibgp *bool `json:"ibgp"`

	// Communities per IP address range. Field introduced in 17.1.3. Maximum of 1024 items allowed.
	IPCommunities []*IPCommunity `json:"ip_communities,omitempty"`

	// Keepalive interval for Peers. Allowed values are 0-3600.
	KeepaliveInterval *int32 `json:"keepalive_interval,omitempty"`

	// Local Autonomous System ID. Allowed values are 1-4294967295.
	// Required: true
	LocalAs *int32 `json:"local_as"`

	// LOCAL_PREF to be used for routes advertised. Applicable only over iBGP. Field introduced in 20.1.1.
	LocalPreference *int32 `json:"local_preference,omitempty"`

	// Number of times the local AS should be prepended additionally. Allowed values are 1-10. Field introduced in 20.1.1.
	NumAsPathPrepend *int32 `json:"num_as_path_prepend,omitempty"`

	// BGP Peers. Maximum of 128 items allowed.
	Peers []*BgpPeer `json:"peers,omitempty"`

	// Learning and advertising options for BGP peers. Field introduced in 20.1.1. Maximum of 128 items allowed.
	RoutingOptions []*BgpRoutingOptions `json:"routing_options,omitempty"`

	// Send community attribute to all peers. Field introduced in 17.1.2.
	SendCommunity *bool `json:"send_community,omitempty"`

	// Shutdown the bgp. Field introduced in 17.2.4.
	Shutdown *bool `json:"shutdown,omitempty"`
}
