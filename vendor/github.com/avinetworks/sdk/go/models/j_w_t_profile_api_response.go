package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JWTProfileAPIResponse j w t profile Api response
// swagger:model JWTProfileApiResponse
type JWTProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*JWTProfile `json:"results,omitempty"`
}
