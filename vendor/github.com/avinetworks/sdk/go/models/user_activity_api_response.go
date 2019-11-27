package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UserActivityAPIResponse user activity Api response
// swagger:model UserActivityApiResponse
type UserActivityAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*UserActivity `json:"results,omitempty"`
}
