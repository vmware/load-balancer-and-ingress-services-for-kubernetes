package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VirtualServiceAPIResponse virtual service Api response
// swagger:model VirtualServiceApiResponse
type VirtualServiceAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VirtualService `json:"results,omitempty"`
}
