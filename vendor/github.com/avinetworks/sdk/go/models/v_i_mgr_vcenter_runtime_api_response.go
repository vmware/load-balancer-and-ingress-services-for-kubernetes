package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrVcenterRuntimeAPIResponse v i mgr vcenter runtime Api response
// swagger:model VIMgrVcenterRuntimeApiResponse
type VIMgrVcenterRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIMgrVcenterRuntime `json:"results,omitempty"`
}
