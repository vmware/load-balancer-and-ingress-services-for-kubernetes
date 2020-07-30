package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServiceEngineGroupAPIResponse service engine group Api response
// swagger:model ServiceEngineGroupApiResponse
type ServiceEngineGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServiceEngineGroup `json:"results,omitempty"`
}
