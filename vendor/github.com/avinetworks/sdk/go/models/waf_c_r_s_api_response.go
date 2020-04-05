package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafCRSAPIResponse waf c r s Api response
// swagger:model WafCRSApiResponse
type WafCRSAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*WafCRS `json:"results,omitempty"`
}
