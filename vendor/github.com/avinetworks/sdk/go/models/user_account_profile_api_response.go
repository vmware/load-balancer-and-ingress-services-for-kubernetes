package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UserAccountProfileAPIResponse user account profile Api response
// swagger:model UserAccountProfileApiResponse
type UserAccountProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*UserAccountProfile `json:"results,omitempty"`
}
