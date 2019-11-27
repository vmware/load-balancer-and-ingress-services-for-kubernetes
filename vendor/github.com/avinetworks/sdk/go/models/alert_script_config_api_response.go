package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertScriptConfigAPIResponse alert script config Api response
// swagger:model AlertScriptConfigApiResponse
type AlertScriptConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AlertScriptConfig `json:"results,omitempty"`
}
