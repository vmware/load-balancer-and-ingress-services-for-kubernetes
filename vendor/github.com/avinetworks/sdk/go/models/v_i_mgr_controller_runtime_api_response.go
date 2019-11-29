package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrControllerRuntimeAPIResponse v i mgr controller runtime Api response
// swagger:model VIMgrControllerRuntimeApiResponse
type VIMgrControllerRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIMgrControllerRuntime `json:"results,omitempty"`
}
