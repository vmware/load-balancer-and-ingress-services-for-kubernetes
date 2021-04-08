package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthProfileAPIResponse auth profile Api response
// swagger:model AuthProfileApiResponse
type AuthProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AuthProfile `json:"results,omitempty"`
}
