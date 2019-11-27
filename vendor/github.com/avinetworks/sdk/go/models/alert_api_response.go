package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertAPIResponse alert Api response
// swagger:model AlertApiResponse
type AlertAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Alert `json:"results,omitempty"`
}
