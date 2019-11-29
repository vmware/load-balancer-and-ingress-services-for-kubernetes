package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkProfileUnion network profile union
// swagger:model NetworkProfileUnion
type NetworkProfileUnion struct {

	// Placeholder for description of property tcp_fast_path_profile of obj type NetworkProfileUnion field type str  type object
	TCPFastPathProfile *TCPFastPathProfile `json:"tcp_fast_path_profile,omitempty"`

	// Placeholder for description of property tcp_proxy_profile of obj type NetworkProfileUnion field type str  type object
	TCPProxyProfile *TCPProxyProfile `json:"tcp_proxy_profile,omitempty"`

	// Configure one of either proxy or fast path profiles. Enum options - PROTOCOL_TYPE_TCP_PROXY, PROTOCOL_TYPE_TCP_FAST_PATH, PROTOCOL_TYPE_UDP_FAST_PATH, PROTOCOL_TYPE_UDP_PROXY.
	// Required: true
	Type *string `json:"type"`

	// Placeholder for description of property udp_fast_path_profile of obj type NetworkProfileUnion field type str  type object
	UDPFastPathProfile *UDPFastPathProfile `json:"udp_fast_path_profile,omitempty"`

	// Configure UDP Proxy network profile. Field introduced in 17.2.8, 18.1.3, 18.2.1.
	UDPProxyProfile *UDPProxyProfile `json:"udp_proxy_profile,omitempty"`
}
