package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GeoDBAPIResponse geo d b Api response
// swagger:model GeoDBApiResponse
type GeoDBAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*GeoDB `json:"results,omitempty"`
}
