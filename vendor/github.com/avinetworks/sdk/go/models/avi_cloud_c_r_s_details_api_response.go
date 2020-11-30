package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AviCloudCRSDetailsAPIResponse avi cloud c r s details Api response
// swagger:model AviCloudCRSDetailsApiResponse
type AviCloudCRSDetailsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AviCloudCRSDetails `json:"results,omitempty"`
}
