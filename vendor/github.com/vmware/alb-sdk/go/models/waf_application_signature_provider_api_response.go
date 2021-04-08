package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafApplicationSignatureProviderAPIResponse waf application signature provider Api response
// swagger:model WafApplicationSignatureProviderApiResponse
type WafApplicationSignatureProviderAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*WafApplicationSignatureProvider `json:"results,omitempty"`
}
