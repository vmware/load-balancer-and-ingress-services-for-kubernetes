package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolAPIResponse pool Api response
// swagger:model PoolApiResponse
type PoolAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Pool `json:"results,omitempty"`
}
