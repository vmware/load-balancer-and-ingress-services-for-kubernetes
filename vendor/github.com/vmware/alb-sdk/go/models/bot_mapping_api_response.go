package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotMappingAPIResponse bot mapping Api response
// swagger:model BotMappingApiResponse
type BotMappingAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BotMapping `json:"results,omitempty"`
}
