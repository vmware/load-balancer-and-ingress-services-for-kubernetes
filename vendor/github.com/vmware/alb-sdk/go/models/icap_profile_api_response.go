package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IcapProfileAPIResponse icap profile Api response
// swagger:model IcapProfileApiResponse
type IcapProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*IcapProfile `json:"results,omitempty"`
}
