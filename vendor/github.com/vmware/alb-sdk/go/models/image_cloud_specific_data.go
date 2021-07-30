// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ImageCloudSpecificData image cloud specific data
// swagger:model ImageCloudSpecificData
type ImageCloudSpecificData struct {

	// Each cloud has specific parameters. Field introduced in 20.1.1.
	Key *string `json:"key,omitempty"`

	// Each parameter can have multiple values. Field introduced in 20.1.1.
	Values []string `json:"values,omitempty"`
}
