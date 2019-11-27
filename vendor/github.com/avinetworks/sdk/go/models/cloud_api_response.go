package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudAPIResponse cloud Api response
// swagger:model CloudApiResponse
type CloudAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Cloud `json:"results,omitempty"`
}
