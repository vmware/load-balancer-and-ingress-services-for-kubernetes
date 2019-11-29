package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TCPApplicationProfile TCP application profile
// swagger:model TCPApplicationProfile
type TCPApplicationProfile struct {

	// Enable/Disable the usage of proxy protocol to convey client connection information to the back-end servers.  Valid only for L4 application profiles and TCP proxy.
	ProxyProtocolEnabled *bool `json:"proxy_protocol_enabled,omitempty"`

	// Version of proxy protocol to be used to convey client connection information to the back-end servers. Enum options - PROXY_PROTOCOL_VERSION_1, PROXY_PROTOCOL_VERSION_2.
	ProxyProtocolVersion *string `json:"proxy_protocol_version,omitempty"`
}
