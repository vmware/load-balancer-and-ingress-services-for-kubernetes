package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PingAccessAgentAPIResponse ping access agent Api response
// swagger:model PingAccessAgentApiResponse
type PingAccessAgentAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*PingAccessAgent `json:"results,omitempty"`
}
