package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TestSeDatastoreLevel1APIResponse test se datastore level1 Api response
// swagger:model TestSeDatastoreLevel1ApiResponse
type TestSeDatastoreLevel1APIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TestSeDatastoreLevel1 `json:"results,omitempty"`
}
