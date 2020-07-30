package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecurityPolicyAPIResponse security policy Api response
// swagger:model SecurityPolicyApiResponse
type SecurityPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SecurityPolicy `json:"results,omitempty"`
}
