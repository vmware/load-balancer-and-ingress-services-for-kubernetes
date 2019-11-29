package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugServiceEngineAPIResponse debug service engine Api response
// swagger:model DebugServiceEngineApiResponse
type DebugServiceEngineAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*DebugServiceEngine `json:"results,omitempty"`
}
