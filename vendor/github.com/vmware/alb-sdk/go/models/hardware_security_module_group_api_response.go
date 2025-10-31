// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HardwareSecurityModuleGroupAPIResponse hardware security module group Api response
// swagger:model HardwareSecurityModuleGroupApiResponse
type HardwareSecurityModuleGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*HardwareSecurityModuleGroup `json:"results,omitempty"`
}
