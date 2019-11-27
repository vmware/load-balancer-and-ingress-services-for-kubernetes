package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SePropertiesAPIResponse se properties Api response
// swagger:model SePropertiesApiResponse
type SePropertiesAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SeProperties `json:"results,omitempty"`
}
