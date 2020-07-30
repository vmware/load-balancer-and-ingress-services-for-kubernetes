package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecurityManagerDataAPIResponse security manager data Api response
// swagger:model SecurityManagerDataApiResponse
type SecurityManagerDataAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SecurityManagerData `json:"results,omitempty"`
}
