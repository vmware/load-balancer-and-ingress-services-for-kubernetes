package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SCVsStateInfoAPIResponse s c vs state info Api response
// swagger:model SCVsStateInfoApiResponse
type SCVsStateInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SCVsStateInfo `json:"results,omitempty"`
}
