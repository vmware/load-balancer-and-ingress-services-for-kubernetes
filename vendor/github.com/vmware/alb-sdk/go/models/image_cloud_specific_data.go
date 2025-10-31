// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ImageCloudSpecificData image cloud specific data
// swagger:model ImageCloudSpecificData
type ImageCloudSpecificData struct {

	// Each cloud has specific parameters. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Key *string `json:"key,omitempty"`

	// Each parameter can have multiple values. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Values []string `json:"values,omitempty"`
}
