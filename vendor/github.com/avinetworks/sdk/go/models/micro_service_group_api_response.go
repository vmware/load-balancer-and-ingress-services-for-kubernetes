package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MicroServiceGroupAPIResponse micro service group Api response
// swagger:model MicroServiceGroupApiResponse
type MicroServiceGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*MicroServiceGroup `json:"results,omitempty"`
}
