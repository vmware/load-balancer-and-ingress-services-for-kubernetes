package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPCommunity Ip community
// swagger:model IpCommunity
type IPCommunity struct {

	// Community *string either in aa nn format where aa, nn is within [1,65535] or local-AS|no-advertise|no-export|internet. Field introduced in 17.1.3.
	Community []string `json:"community,omitempty"`

	// Beginning of IP address range. Field introduced in 17.1.3.
	// Required: true
	IPBegin *IPAddr `json:"ip_begin"`

	// End of IP address range. Optional if ip_begin is the only IP address in specified IP range. Field introduced in 17.1.3.
	IPEnd *IPAddr `json:"ip_end,omitempty"`
}
