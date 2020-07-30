package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BfdProfile bfd profile
// swagger:model BfdProfile
type BfdProfile struct {

	// Default required minimum receive interval (in ms) used in BFD. Allowed values are 500-4000000. Field introduced in 20.1.1.
	Minrx *int32 `json:"minrx,omitempty"`

	// Default desired minimum transmit interval (in ms) used in BFD. Allowed values are 500-4000000. Field introduced in 20.1.1.
	Mintx *int32 `json:"mintx,omitempty"`

	// Default detection multiplier used in BFD. Allowed values are 3-255. Field introduced in 20.1.1.
	Multi *int32 `json:"multi,omitempty"`
}
