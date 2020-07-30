package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CRSDetailsAPIResponse c r s details Api response
// swagger:model CRSDetailsApiResponse
type CRSDetailsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*CRSDetails `json:"results,omitempty"`
}
