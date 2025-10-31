// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ImageCloudData image cloud data
// swagger:model ImageCloudData
type ImageCloudData struct {

	// Cloud Data specific to a particular cloud. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudDataValues []*ImageCloudSpecificData `json:"cloud_data_values,omitempty"`

	// Contains the name of the cloud. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudName *string `json:"cloud_name,omitempty"`
}
