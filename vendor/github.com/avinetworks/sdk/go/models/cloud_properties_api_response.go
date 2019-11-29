package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudPropertiesAPIResponse cloud properties Api response
// swagger:model CloudPropertiesApiResponse
type CloudPropertiesAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*CloudProperties `json:"results,omitempty"`
}
