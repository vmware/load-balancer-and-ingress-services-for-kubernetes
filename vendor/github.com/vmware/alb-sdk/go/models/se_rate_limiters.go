package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeRateLimiters se rate limiters
// swagger:model SeRateLimiters
type SeRateLimiters struct {

	// Rate limiter for ARP packets in pps.
	ArpRl *int32 `json:"arp_rl,omitempty"`

	// Default Rate limiter in pps.
	DefaultRl *int32 `json:"default_rl,omitempty"`

	// Rate limiter for number of flow probes in pps.
	FlowProbeRl *int32 `json:"flow_probe_rl,omitempty"`

	// Rate limiter for ICMP requests in pps.
	IcmpRl *int32 `json:"icmp_rl,omitempty"`

	// Rate limiter for ICMP response in pps.
	IcmpRspRl *int32 `json:"icmp_rsp_rl,omitempty"`

	// Rate limiter for number RST pkts sent in pps.
	RstRl *int32 `json:"rst_rl,omitempty"`
}
