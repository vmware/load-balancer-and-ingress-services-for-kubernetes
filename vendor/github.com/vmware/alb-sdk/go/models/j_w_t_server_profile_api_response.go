package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JWTServerProfileAPIResponse j w t server profile Api response
// swagger:model JWTServerProfileApiResponse
type JWTServerProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*JWTServerProfile `json:"results,omitempty"`
}
