package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrNWRuntimeAPIResponse v i mgr n w runtime Api response
// swagger:model VIMgrNWRuntimeApiResponse
type VIMgrNWRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIMgrNWRuntime `json:"results,omitempty"`
}
