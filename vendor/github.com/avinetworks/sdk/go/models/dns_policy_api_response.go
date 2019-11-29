package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSPolicyAPIResponse Dns policy Api response
// swagger:model DnsPolicyApiResponse
type DNSPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*DNSPolicy `json:"results,omitempty"`
}
