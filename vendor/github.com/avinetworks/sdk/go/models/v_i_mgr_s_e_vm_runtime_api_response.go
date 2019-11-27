package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrSEVMRuntimeAPIResponse v i mgr s e VM runtime Api response
// swagger:model VIMgrSEVMRuntimeApiResponse
type VIMgrSEVMRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIMgrSEVMRuntime `json:"results,omitempty"`
}
