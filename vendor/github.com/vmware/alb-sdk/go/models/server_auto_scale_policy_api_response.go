package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerAutoScalePolicyAPIResponse server auto scale policy Api response
// swagger:model ServerAutoScalePolicyApiResponse
type ServerAutoScalePolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServerAutoScalePolicy `json:"results,omitempty"`
}
