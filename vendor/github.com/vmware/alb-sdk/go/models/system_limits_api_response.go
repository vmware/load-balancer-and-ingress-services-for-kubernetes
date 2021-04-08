package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SystemLimitsAPIResponse system limits Api response
// swagger:model SystemLimitsApiResponse
type SystemLimitsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SystemLimits `json:"results,omitempty"`
}
