package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FileObjectAPIResponse file object Api response
// swagger:model FileObjectApiResponse
type FileObjectAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*FileObject `json:"results,omitempty"`
}
