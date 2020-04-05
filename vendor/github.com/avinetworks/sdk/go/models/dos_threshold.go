package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DosThreshold dos threshold
// swagger:model DosThreshold
type DosThreshold struct {

	// Attack type. Enum options - LAND, SMURF, ICMP_PING_FLOOD, UNKOWN_PROTOCOL, TEARDROP, IP_FRAG_OVERRUN, IP_FRAG_TOOSMALL, IP_FRAG_FULL, IP_FRAG_INCOMPLETE, PORT_SCAN, TCP_NON_SYN_FLOOD_OLD, SYN_FLOOD, BAD_RST_FLOOD, MALFORMED_FLOOD, FAKE_SESSION, ZERO_WINDOW_STRESS, SMALL_WINDOW_STRESS, DOS_HTTP_TIMEOUT, DOS_HTTP_ERROR, DOS_HTTP_ABORT...
	// Required: true
	Attack *string `json:"attack"`

	// Maximum number of packets or connections or requests in a given interval of time to be deemed as attack.
	// Required: true
	MaxValue *int32 `json:"max_value"`

	// Minimum number of packets or connections or requests in a given interval of time to be deemed as attack.
	// Required: true
	MinValue *int32 `json:"min_value"`
}
