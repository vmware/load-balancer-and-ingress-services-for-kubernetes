package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbAPIResponse gslb Api response
// swagger:model GslbApiResponse
type GslbAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Gslb `json:"results,omitempty"`
}
