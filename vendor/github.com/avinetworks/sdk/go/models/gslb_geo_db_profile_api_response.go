package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbGeoDbProfileAPIResponse gslb geo db profile Api response
// swagger:model GslbGeoDbProfileApiResponse
type GslbGeoDbProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*GslbGeoDbProfile `json:"results,omitempty"`
}
