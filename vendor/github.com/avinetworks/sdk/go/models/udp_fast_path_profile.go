package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UDPFastPathProfile UDP fast path profile
// swagger:model UDPFastPathProfile
type UDPFastPathProfile struct {

	// DSR profile information. Field introduced in 18.2.3.
	DsrProfile *DsrProfile `json:"dsr_profile,omitempty"`

	// When enabled, every UDP packet is considered a new transaction and may be load balanced to a different server.  When disabled, packets from the same client source IP and port are sent to the same server.
	PerPktLoadbalance *bool `json:"per_pkt_loadbalance,omitempty"`

	// The amount of time (in sec) for which a flow needs to be idle before it is deleted. Allowed values are 2-3600.
	SessionIDLETimeout *int32 `json:"session_idle_timeout,omitempty"`

	// When disabled, Source NAT will not be performed for all client UDP packets.
	Snat *bool `json:"snat,omitempty"`
}
