package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbIPAddr gslb Ip addr
// swagger:model GslbIpAddr
type GslbIPAddr struct {

	// Public IP address of the pool member. Field introduced in 17.1.2.
	IP *IPAddr `json:"ip,omitempty"`
}
