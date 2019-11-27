package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerPropertiesAPIResponse controller properties Api response
// swagger:model ControllerPropertiesApiResponse
type ControllerPropertiesAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ControllerProperties `json:"results,omitempty"`
}
