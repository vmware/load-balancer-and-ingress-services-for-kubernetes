// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerPropertiesAPIResponse controller properties Api response
// swagger:model ControllerPropertiesApiResponse
type ControllerPropertiesAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ControllerProperties `json:"results,omitempty"`
}
