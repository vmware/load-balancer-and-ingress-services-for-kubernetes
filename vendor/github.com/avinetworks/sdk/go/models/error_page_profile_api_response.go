package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ErrorPageProfileAPIResponse error page profile Api response
// swagger:model ErrorPageProfileApiResponse
type ErrorPageProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ErrorPageProfile `json:"results,omitempty"`
}
