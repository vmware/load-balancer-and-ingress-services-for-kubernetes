package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FederationCheckpointAPIResponse federation checkpoint Api response
// swagger:model FederationCheckpointApiResponse
type FederationCheckpointAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*FederationCheckpoint `json:"results,omitempty"`
}
