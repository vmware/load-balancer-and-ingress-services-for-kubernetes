package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ActionGroupConfigAPIResponse action group config Api response
// swagger:model ActionGroupConfigApiResponse
type ActionGroupConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ActionGroupConfig `json:"results,omitempty"`
}
