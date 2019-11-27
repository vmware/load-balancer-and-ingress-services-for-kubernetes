package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudRuntimeAPIResponse cloud runtime Api response
// swagger:model CloudRuntimeApiResponse
type CloudRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*CloudRuntime `json:"results,omitempty"`
}
