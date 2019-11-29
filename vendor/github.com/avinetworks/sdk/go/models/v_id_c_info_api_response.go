package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIDCInfoAPIResponse v ID c info Api response
// swagger:model VIDCInfoApiResponse
type VIDCInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIDCInfo `json:"results,omitempty"`
}
