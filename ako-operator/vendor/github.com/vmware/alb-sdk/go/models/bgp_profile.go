// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BgpProfile bgp profile
// swagger:model BgpProfile
type BgpProfile struct {

	// Community *string either in aa nn format where aa, nn is within [1,65535] or local-AS|no-advertise|no-export|internet. Field introduced in 17.1.2. Maximum of 16 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Community []string `json:"community,omitempty"`

	// Hold time for Peers. Allowed values are 3-7200. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HoldTime *uint32 `json:"hold_time,omitempty"`

	// BGP peer type. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Ibgp *bool `json:"ibgp"`

	// Communities per IP address range. Field introduced in 17.1.3. Maximum of 1024 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPCommunities []*IPCommunity `json:"ip_communities,omitempty"`

	// Keepalive interval for Peers. Allowed values are 0-3600. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeepaliveInterval *uint32 `json:"keepalive_interval,omitempty"`

	// Local Autonomous System ID. Allowed values are 1-4294967295. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	LocalAs *uint32 `json:"local_as"`

	// LOCAL_PREF to be used for routes advertised. Applicable only over iBGP. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LocalPreference *uint32 `json:"local_preference,omitempty"`

	// Number of times the local AS should be prepended additionally. Allowed values are 1-10. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumAsPathPrepend *uint32 `json:"num_as_path_prepend,omitempty"`

	// BGP Peers. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Peers []*BgpPeer `json:"peers,omitempty"`

	// Learning and advertising options for BGP peers. Field introduced in 20.1.1. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RoutingOptions []*BgpRoutingOptions `json:"routing_options,omitempty"`

	// Send community attribute to all peers. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SendCommunity *bool `json:"send_community,omitempty"`

	// Shutdown the bgp. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Shutdown *bool `json:"shutdown,omitempty"`
}
