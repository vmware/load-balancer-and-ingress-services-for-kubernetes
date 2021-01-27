package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TCPFastPathProfile TCP fast path profile
// swagger:model TCPFastPathProfile
type TCPFastPathProfile struct {

	// DSR profile information. Field introduced in 18.2.3. Allowed in Basic edition, Essentials edition, Enterprise edition.
	DsrProfile *DsrProfile `json:"dsr_profile,omitempty"`

	// When enabled, Avi will complete the 3-way handshake with the client before forwarding any packets to the server.  This will protect the server from SYN flood and half open SYN connections. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	EnableSynProtection *bool `json:"enable_syn_protection,omitempty"`

	// The amount of time (in sec) for which a connection needs to be idle before it is eligible to be deleted. Allowed values are 5-14400. Special values are 0 - 'infinite'. Unit is SEC.
	SessionIDLETimeout *int32 `json:"session_idle_timeout,omitempty"`
}
