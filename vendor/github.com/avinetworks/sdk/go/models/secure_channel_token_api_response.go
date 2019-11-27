package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecureChannelTokenAPIResponse secure channel token Api response
// swagger:model SecureChannelTokenApiResponse
type SecureChannelTokenAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SecureChannelToken `json:"results,omitempty"`
}
