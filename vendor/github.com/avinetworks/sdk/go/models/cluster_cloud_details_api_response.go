package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterCloudDetailsAPIResponse cluster cloud details Api response
// swagger:model ClusterCloudDetailsApiResponse
type ClusterCloudDetailsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ClusterCloudDetails `json:"results,omitempty"`
}
