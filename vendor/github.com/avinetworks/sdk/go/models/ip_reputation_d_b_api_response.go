package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPReputationDBAPIResponse IP reputation d b Api response
// swagger:model IPReputationDBApiResponse
type IPReputationDBAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*IPReputationDB `json:"results,omitempty"`
}
