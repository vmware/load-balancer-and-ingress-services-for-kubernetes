package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VCenterServerAPIResponse v center server Api response
// swagger:model VCenterServerApiResponse
type VCenterServerAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VCenterServer `json:"results,omitempty"`
}
