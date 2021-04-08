package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkRuntimeAPIResponse network runtime Api response
// swagger:model NetworkRuntimeApiResponse
type NetworkRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NetworkRuntime `json:"results,omitempty"`
}
