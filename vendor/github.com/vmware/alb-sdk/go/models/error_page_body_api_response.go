// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ErrorPageBodyAPIResponse error page body Api response
// swagger:model ErrorPageBodyApiResponse
type ErrorPageBodyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ErrorPageBody `json:"results,omitempty"`
}
