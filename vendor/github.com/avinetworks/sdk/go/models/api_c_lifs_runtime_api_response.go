package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// APICLifsRuntimeAPIResponse API c lifs runtime Api response
// swagger:model APICLifsRuntimeApiResponse
type APICLifsRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*APICLifsRuntime `json:"results,omitempty"`
}
