package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafProfileAPIResponse waf profile Api response
// swagger:model WafProfileApiResponse
type WafProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*WafProfile `json:"results,omitempty"`
}
