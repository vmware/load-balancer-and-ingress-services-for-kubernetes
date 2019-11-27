package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertObjectListAPIResponse alert object list Api response
// swagger:model AlertObjectListApiResponse
type AlertObjectListAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AlertObjectList `json:"results,omitempty"`
}
