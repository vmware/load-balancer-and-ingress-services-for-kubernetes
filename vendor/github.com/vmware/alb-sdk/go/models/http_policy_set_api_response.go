package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPPolicySetAPIResponse HTTP policy set Api response
// swagger:model HTTPPolicySetApiResponse
type HTTPPolicySetAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*HTTPPolicySet `json:"results,omitempty"`
}
