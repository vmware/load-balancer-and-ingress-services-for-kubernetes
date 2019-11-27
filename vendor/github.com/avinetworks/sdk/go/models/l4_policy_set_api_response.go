package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4PolicySetAPIResponse l4 policy set Api response
// swagger:model L4PolicySetApiResponse
type L4PolicySetAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*L4PolicySet `json:"results,omitempty"`
}
