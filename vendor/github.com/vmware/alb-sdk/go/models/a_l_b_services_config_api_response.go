package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ALBServicesConfigAPIResponse a l b services config Api response
// swagger:model ALBServicesConfigApiResponse
type ALBServicesConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ALBServicesConfig `json:"results,omitempty"`
}
