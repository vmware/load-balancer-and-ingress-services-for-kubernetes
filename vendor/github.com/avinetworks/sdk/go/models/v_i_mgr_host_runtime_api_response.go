package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrHostRuntimeAPIResponse v i mgr host runtime Api response
// swagger:model VIMgrHostRuntimeApiResponse
type VIMgrHostRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIMgrHostRuntime `json:"results,omitempty"`
}
