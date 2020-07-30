package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkAPIResponse network Api response
// swagger:model NetworkApiResponse
type NetworkAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*Network `json:"results,omitempty"`
}
