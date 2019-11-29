package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PriorityLabelsAPIResponse priority labels Api response
// swagger:model PriorityLabelsApiResponse
type PriorityLabelsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*PriorityLabels `json:"results,omitempty"`
}
