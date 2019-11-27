package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyAPIResponse waf policy Api response
// swagger:model WafPolicyApiResponse
type WafPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*WafPolicy `json:"results,omitempty"`
}
