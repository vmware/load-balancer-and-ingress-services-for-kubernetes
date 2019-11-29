package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLProfileAPIResponse s s l profile Api response
// swagger:model SSLProfileApiResponse
type SSLProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SSLProfile `json:"results,omitempty"`
}
