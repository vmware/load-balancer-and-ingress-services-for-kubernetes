package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerLicenseAPIResponse controller license Api response
// swagger:model ControllerLicenseApiResponse
type ControllerLicenseAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ControllerLicense `json:"results,omitempty"`
}
