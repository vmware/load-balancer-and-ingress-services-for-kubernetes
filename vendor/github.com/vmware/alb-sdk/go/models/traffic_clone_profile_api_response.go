package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TrafficCloneProfileAPIResponse traffic clone profile Api response
// swagger:model TrafficCloneProfileApiResponse
type TrafficCloneProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TrafficCloneProfile `json:"results,omitempty"`
}
