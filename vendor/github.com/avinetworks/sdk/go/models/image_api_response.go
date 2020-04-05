package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ImageAPIResponse image Api response
// swagger:model ImageApiResponse
type ImageAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Image `json:"results,omitempty"`
}
