// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SystemConfigurationAPIResponse system configuration Api response
// swagger:model SystemConfigurationApiResponse
type SystemConfigurationAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SystemConfiguration `json:"results,omitempty"`
}
