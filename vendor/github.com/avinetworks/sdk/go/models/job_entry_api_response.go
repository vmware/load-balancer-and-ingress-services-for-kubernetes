package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JobEntryAPIResponse job entry Api response
// swagger:model JobEntryApiResponse
type JobEntryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*JobEntry `json:"results,omitempty"`
}
