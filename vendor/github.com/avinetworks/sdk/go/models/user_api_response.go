package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UserAPIResponse user Api response
// swagger:model UserApiResponse
type UserAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*User `json:"results,omitempty"`
}
