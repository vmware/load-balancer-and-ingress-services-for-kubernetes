// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BgpPeer bgp peer
// swagger:model BgpPeer
type BgpPeer struct {

	// Advertise SNAT IP to this Peer. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdvertiseSnatIP *bool `json:"advertise_snat_ip,omitempty"`

	// Advertise VIP to this Peer. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdvertiseVip *bool `json:"advertise_vip,omitempty"`

	// Advertisement interval for this Peer. Allowed values are 1-60. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdvertisementInterval *uint32 `json:"advertisement_interval,omitempty"`

	// Enable Bi-Directional Forward Detection. Only async mode supported. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Bfd *bool `json:"bfd,omitempty"`

	// Connect timer for this Peer. Allowed values are 1-120. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnectTimer *uint32 `json:"connect_timer,omitempty"`

	// TTL for multihop ebgp Peer. Allowed values are 0-255. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EbgpMultihop *uint32 `json:"ebgp_multihop,omitempty"`

	// Hold time for this Peer. Allowed values are 3-7200. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HoldTime *uint32 `json:"hold_time,omitempty"`

	// Override the profile level local_as with the peer level remote_as. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IbgpLocalAsOverride *bool `json:"ibgp_local_as_override,omitempty"`

	// Keepalive interval for this Peer. Allowed values are 0-3600. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeepaliveInterval *uint32 `json:"keepalive_interval,omitempty"`

	// Label used to enable learning and/or advertisement of routes to this peer. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Label *string `json:"label,omitempty"`

	// Local AS to use for this ebgp peer. If specified, this will override the local AS configured at the VRF level. Allowed values are 1-4294967295. Field introduced in 17.1.6,17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LocalAs *uint32 `json:"local_as,omitempty"`

	// Peer Autonomous System Md5 Digest Secret Key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Md5Secret *string `json:"md5_secret,omitempty"`

	// Network providing reachability for Peer. It is a reference to an object of type Network. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkRef *string `json:"network_ref,omitempty"`

	// IP Address of the BGP Peer. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PeerIP *IPAddr `json:"peer_ip,omitempty"`

	// IPv6 Address of the BGP Peer. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PeerIp6 *IPAddr `json:"peer_ip6,omitempty"`

	// Peer Autonomous System ID. Allowed values are 1-4294967295. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RemoteAs *uint32 `json:"remote_as,omitempty"`

	// Shutdown the BGP peer. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Shutdown *bool `json:"shutdown,omitempty"`

	// Subnet providing reachability for Peer. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// IPv6 subnet providing reachability for Peer. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Subnet6 *IPAddrPrefix `json:"subnet6,omitempty"`
}
