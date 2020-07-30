package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SystemConfigurationAPIResponse system configuration Api response
// swagger:model SystemConfigurationApiResponse
type SystemConfigurationAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SystemConfiguration `json:"results,omitempty"`
}
