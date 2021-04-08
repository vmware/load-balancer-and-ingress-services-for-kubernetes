package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSOPolicyAPIResponse s s o policy Api response
// swagger:model SSOPolicyApiResponse
type SSOPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SSOPolicy `json:"results,omitempty"`
}
