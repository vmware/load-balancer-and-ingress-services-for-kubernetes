package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SiteVersionAPIResponse site version Api response
// swagger:model SiteVersionApiResponse
type SiteVersionAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SiteVersion `json:"results,omitempty"`
}
