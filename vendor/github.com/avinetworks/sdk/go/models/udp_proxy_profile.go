package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UDPProxyProfile UDP proxy profile
// swagger:model UDPProxyProfile
type UDPProxyProfile struct {

	// The amount of time (in sec) for which a flow needs to be idle before it is deleted. Allowed values are 2-3600. Field introduced in 17.2.8, 18.1.3, 18.2.1.
	SessionIDLETimeout *int32 `json:"session_idle_timeout,omitempty"`
}
