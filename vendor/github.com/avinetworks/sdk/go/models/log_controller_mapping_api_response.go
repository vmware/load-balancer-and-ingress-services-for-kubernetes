package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LogControllerMappingAPIResponse log controller mapping Api response
// swagger:model LogControllerMappingApiResponse
type LogControllerMappingAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*LogControllerMapping `json:"results,omitempty"`
}
