package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ObjectAccessPolicyAPIResponse object access policy Api response
// swagger:model ObjectAccessPolicyApiResponse
type ObjectAccessPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ObjectAccessPolicy `json:"results,omitempty"`
}
