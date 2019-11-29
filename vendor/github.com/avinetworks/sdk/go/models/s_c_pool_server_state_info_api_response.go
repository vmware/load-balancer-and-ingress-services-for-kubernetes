package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SCPoolServerStateInfoAPIResponse s c pool server state info Api response
// swagger:model SCPoolServerStateInfoApiResponse
type SCPoolServerStateInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SCPoolServerStateInfo `json:"results,omitempty"`
}
