package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyPSMGroupAPIResponse waf policy p s m group Api response
// swagger:model WafPolicyPSMGroupApiResponse
type WafPolicyPSMGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*WafPolicyPSMGroup `json:"results,omitempty"`
}
