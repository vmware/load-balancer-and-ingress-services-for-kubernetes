package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIPGNameInfoAPIResponse v IP g name info Api response
// swagger:model VIPGNameInfoApiResponse
type VIPGNameInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VIPGNameInfo `json:"results,omitempty"`
}
