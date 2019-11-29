package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertConfigAPIResponse alert config Api response
// swagger:model AlertConfigApiResponse
type AlertConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AlertConfig `json:"results,omitempty"`
}
