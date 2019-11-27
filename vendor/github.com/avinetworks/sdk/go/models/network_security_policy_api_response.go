package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkSecurityPolicyAPIResponse network security policy Api response
// swagger:model NetworkSecurityPolicyApiResponse
type NetworkSecurityPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*NetworkSecurityPolicy `json:"results,omitempty"`
}
