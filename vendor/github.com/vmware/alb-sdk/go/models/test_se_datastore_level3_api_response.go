package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TestSeDatastoreLevel3APIResponse test se datastore level3 Api response
// swagger:model TestSeDatastoreLevel3ApiResponse
type TestSeDatastoreLevel3APIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TestSeDatastoreLevel3 `json:"results,omitempty"`
}
