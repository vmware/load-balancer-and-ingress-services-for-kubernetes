package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApplicationAPIResponse application Api response
// swagger:model ApplicationApiResponse
type ApplicationAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Application `json:"results,omitempty"`
}
