package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkProfileAPIResponse network profile Api response
// swagger:model NetworkProfileApiResponse
type NetworkProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*NetworkProfile `json:"results,omitempty"`
}
