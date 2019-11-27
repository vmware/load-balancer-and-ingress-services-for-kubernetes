package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrClusterRuntimeAPIResponse v i mgr cluster runtime Api response
// swagger:model VIMgrClusterRuntimeApiResponse
type VIMgrClusterRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIMgrClusterRuntime `json:"results,omitempty"`
}
