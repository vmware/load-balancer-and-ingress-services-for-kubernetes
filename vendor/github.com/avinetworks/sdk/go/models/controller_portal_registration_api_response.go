package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerPortalRegistrationAPIResponse controller portal registration Api response
// swagger:model ControllerPortalRegistrationApiResponse
type ControllerPortalRegistrationAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ControllerPortalRegistration `json:"results,omitempty"`
}
