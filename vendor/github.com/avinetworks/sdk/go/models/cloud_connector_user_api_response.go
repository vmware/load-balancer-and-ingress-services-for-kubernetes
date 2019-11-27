package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudConnectorUserAPIResponse cloud connector user Api response
// swagger:model CloudConnectorUserApiResponse
type CloudConnectorUserAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*CloudConnectorUser `json:"results,omitempty"`
}
