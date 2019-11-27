package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugControllerAPIResponse debug controller Api response
// swagger:model DebugControllerApiResponse
type DebugControllerAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*DebugController `json:"results,omitempty"`
}
