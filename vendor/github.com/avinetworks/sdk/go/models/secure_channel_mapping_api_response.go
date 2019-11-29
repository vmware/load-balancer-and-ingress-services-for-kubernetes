package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecureChannelMappingAPIResponse secure channel mapping Api response
// swagger:model SecureChannelMappingApiResponse
type SecureChannelMappingAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SecureChannelMapping `json:"results,omitempty"`
}
