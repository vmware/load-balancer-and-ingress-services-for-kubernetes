package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CustomerPortalInfoAPIResponse customer portal info Api response
// swagger:model CustomerPortalInfoApiResponse
type CustomerPortalInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*CustomerPortalInfo `json:"results,omitempty"`
}
