package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterAPIResponse cluster Api response
// swagger:model ClusterApiResponse
type ClusterAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*Cluster `json:"results,omitempty"`
}
