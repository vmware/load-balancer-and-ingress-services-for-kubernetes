package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrVMRuntimeAPIResponse v i mgr VM runtime Api response
// swagger:model VIMgrVMRuntimeApiResponse
type VIMgrVMRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIMgrVMRuntime `json:"results,omitempty"`
}
