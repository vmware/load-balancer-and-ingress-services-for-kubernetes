package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MicroServiceAPIResponse micro service Api response
// swagger:model MicroServiceApiResponse
type MicroServiceAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*MicroService `json:"results,omitempty"`
}
