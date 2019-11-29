package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SidebandProfile sideband profile
// swagger:model SidebandProfile
type SidebandProfile struct {

	// IP Address of the sideband server.
	IP []*IPAddr `json:"ip,omitempty"`

	// Maximum size of the request body that will be sent on the sideband. Allowed values are 0-16384.
	SidebandMaxRequestBodySize *int32 `json:"sideband_max_request_body_size,omitempty"`
}
