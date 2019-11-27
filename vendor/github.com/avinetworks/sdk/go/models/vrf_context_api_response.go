package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VrfContextAPIResponse vrf context Api response
// swagger:model VrfContextApiResponse
type VrfContextAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VrfContext `json:"results,omitempty"`
}
