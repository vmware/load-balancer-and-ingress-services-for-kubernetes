// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecurityManagerDataAPIResponse security manager data Api response
// swagger:model SecurityManagerDataApiResponse
type SecurityManagerDataAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SecurityManagerData `json:"results,omitempty"`
}
