package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ErrorPageBodyAPIResponse error page body Api response
// swagger:model ErrorPageBodyApiResponse
type ErrorPageBodyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ErrorPageBody `json:"results,omitempty"`
}
