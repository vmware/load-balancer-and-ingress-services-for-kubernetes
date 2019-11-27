package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbServiceAPIResponse gslb service Api response
// swagger:model GslbServiceApiResponse
type GslbServiceAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*GslbService `json:"results,omitempty"`
}
