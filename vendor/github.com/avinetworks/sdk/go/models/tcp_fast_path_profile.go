package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TCPFastPathProfile TCP fast path profile
// swagger:model TCPFastPathProfile
type TCPFastPathProfile struct {

	// When enabled, Avi will complete the 3-way handshake with the client before forwarding any packets to the server.  This will protect the server from SYN flood and half open SYN connections.
	EnableSynProtection *bool `json:"enable_syn_protection,omitempty"`

	// The amount of time (in sec) for which a connection needs to be idle before it is eligible to be deleted. Allowed values are 5-3600. Special values are 0 - 'infinite'.
	SessionIDLETimeout *int32 `json:"session_idle_timeout,omitempty"`
}
