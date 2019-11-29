package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugVirtualServiceAPIResponse debug virtual service Api response
// swagger:model DebugVirtualServiceApiResponse
type DebugVirtualServiceAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*DebugVirtualService `json:"results,omitempty"`
}
