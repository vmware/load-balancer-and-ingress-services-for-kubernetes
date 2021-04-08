package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtSegmentRuntimeAPIResponse nsxt segment runtime Api response
// swagger:model NsxtSegmentRuntimeApiResponse
type NsxtSegmentRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NsxtSegmentRuntime `json:"results,omitempty"`
}
