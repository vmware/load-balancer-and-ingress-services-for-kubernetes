// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AutoScaleLaunchConfigAPIResponse auto scale launch config Api response
// swagger:model AutoScaleLaunchConfigApiResponse
type AutoScaleLaunchConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AutoScaleLaunchConfig `json:"results,omitempty"`
}
