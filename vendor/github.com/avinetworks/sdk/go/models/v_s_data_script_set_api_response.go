package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VSDataScriptSetAPIResponse v s data script set Api response
// swagger:model VSDataScriptSetApiResponse
type VSDataScriptSetAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VSDataScriptSet `json:"results,omitempty"`
}
