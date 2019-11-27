package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StringGroupAPIResponse *string group Api response
// swagger:model StringGroupApiResponse
type StringGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*StringGroup `json:"results,omitempty"`
}
