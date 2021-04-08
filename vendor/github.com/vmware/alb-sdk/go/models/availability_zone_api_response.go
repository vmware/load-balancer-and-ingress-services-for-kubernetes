package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AvailabilityZoneAPIResponse availability zone Api response
// swagger:model AvailabilityZoneApiResponse
type AvailabilityZoneAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AvailabilityZone `json:"results,omitempty"`
}
