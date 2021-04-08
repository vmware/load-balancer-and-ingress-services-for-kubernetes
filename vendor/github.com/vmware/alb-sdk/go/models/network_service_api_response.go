package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkServiceAPIResponse network service Api response
// swagger:model NetworkServiceApiResponse
type NetworkServiceAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NetworkService `json:"results,omitempty"`
}
