package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServiceEngineAPIResponse service engine Api response
// swagger:model ServiceEngineApiResponse
type ServiceEngineAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServiceEngine `json:"results,omitempty"`
}
