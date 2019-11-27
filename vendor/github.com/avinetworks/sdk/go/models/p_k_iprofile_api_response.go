package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PKIprofileAPIResponse p k iprofile Api response
// swagger:model PKIProfileApiResponse
type PKIprofileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*PKIprofile `json:"results,omitempty"`
}
