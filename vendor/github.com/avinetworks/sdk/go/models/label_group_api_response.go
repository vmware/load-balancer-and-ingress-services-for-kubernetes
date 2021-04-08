package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LabelGroupAPIResponse label group Api response
// swagger:model LabelGroupApiResponse
type LabelGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*LabelGroup `json:"results,omitempty"`
}
