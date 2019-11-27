package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BgpPeer bgp peer
// swagger:model BgpPeer
type BgpPeer struct {

	// Advertise SNAT IP to this Peer.
	AdvertiseSnatIP *bool `json:"advertise_snat_ip,omitempty"`

	// Advertise VIP to this Peer.
	AdvertiseVip *bool `json:"advertise_vip,omitempty"`

	// Advertisement interval for this Peer. Allowed values are 1-60.
	AdvertisementInterval *int32 `json:"advertisement_interval,omitempty"`

	// Enable Bi-Directional Forward Detection. Only async mode supported.
	Bfd *bool `json:"bfd,omitempty"`

	// Connect timer for this Peer. Allowed values are 1-120.
	ConnectTimer *int32 `json:"connect_timer,omitempty"`

	// TTL for multihop ebgp Peer. Allowed values are 0-255. Field introduced in 17.1.3.
	EbgpMultihop *int32 `json:"ebgp_multihop,omitempty"`

	// Hold time for this Peer. Allowed values are 3-7200.
	HoldTime *int32 `json:"hold_time,omitempty"`

	// Keepalive interval for this Peer. Allowed values are 0-3600.
	KeepaliveInterval *int32 `json:"keepalive_interval,omitempty"`

	// Local AS to use for this ebgp peer. If specified, this will override the local AS configured at the VRF level. Allowed values are 1-4294967295. Field introduced in 17.1.6,17.2.2.
	LocalAs *int32 `json:"local_as,omitempty"`

	// Peer Autonomous System Md5 Digest Secret Key.
	Md5Secret *string `json:"md5_secret,omitempty"`

	// Network providing reachability for Peer. It is a reference to an object of type Network.
	NetworkRef *string `json:"network_ref,omitempty"`

	// IP Address of the BGP Peer.
	PeerIP *IPAddr `json:"peer_ip,omitempty"`

	// IPv6 Address of the BGP Peer. Field introduced in 18.1.1.
	PeerIp6 *IPAddr `json:"peer_ip6,omitempty"`

	// Peer Autonomous System ID. Allowed values are 1-4294967295.
	RemoteAs *int32 `json:"remote_as,omitempty"`

	// Shutdown the BGP peer. Field introduced in 17.2.4.
	Shutdown *bool `json:"shutdown,omitempty"`

	// Subnet providing reachability for Peer.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// IPv6 subnet providing reachability for Peer. Field introduced in 18.1.1.
	Subnet6 *IPAddrPrefix `json:"subnet6,omitempty"`
}
