package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertEmailConfigAPIResponse alert email config Api response
// swagger:model AlertEmailConfigApiResponse
type AlertEmailConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AlertEmailConfig `json:"results,omitempty"`
}
