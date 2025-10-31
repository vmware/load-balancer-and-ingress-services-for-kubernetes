// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineGroupAPIResponse service engine group Api response
// swagger:model ServiceEngineGroupApiResponse
type ServiceEngineGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServiceEngineGroup `json:"results,omitempty"`
}
