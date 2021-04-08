package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ImageCloudData image cloud data
// swagger:model ImageCloudData
type ImageCloudData struct {

	// Cloud Data specific to a particular cloud. Field introduced in 20.1.1.
	CloudDataValues []*ImageCloudSpecificData `json:"cloud_data_values,omitempty"`

	// Contains the name of the cloud. Field introduced in 20.1.1.
	CloudName *string `json:"cloud_name,omitempty"`
}
