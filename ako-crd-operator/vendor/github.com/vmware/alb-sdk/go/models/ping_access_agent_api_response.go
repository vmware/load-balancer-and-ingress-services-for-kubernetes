// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PingAccessAgentAPIResponse ping access agent Api response
// swagger:model PingAccessAgentApiResponse
type PingAccessAgentAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*PingAccessAgent `json:"results,omitempty"`
}
