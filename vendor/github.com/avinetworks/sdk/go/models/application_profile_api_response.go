package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApplicationProfileAPIResponse application profile Api response
// swagger:model ApplicationProfileApiResponse
type ApplicationProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ApplicationProfile `json:"results,omitempty"`
}
