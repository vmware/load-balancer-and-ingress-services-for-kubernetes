package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerSiteAPIResponse controller site Api response
// swagger:model ControllerSiteApiResponse
type ControllerSiteAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ControllerSite `json:"results,omitempty"`
}
