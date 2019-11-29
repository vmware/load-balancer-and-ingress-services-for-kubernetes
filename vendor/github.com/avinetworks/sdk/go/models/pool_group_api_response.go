package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolGroupAPIResponse pool group Api response
// swagger:model PoolGroupApiResponse
type PoolGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*PoolGroup `json:"results,omitempty"`
}
