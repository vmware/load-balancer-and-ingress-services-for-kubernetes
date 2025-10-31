// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineConfigAPIResponse service engine config Api response
// swagger:model ServiceEngineConfigApiResponse
type ServiceEngineConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServiceEngineConfig `json:"results,omitempty"`
}
