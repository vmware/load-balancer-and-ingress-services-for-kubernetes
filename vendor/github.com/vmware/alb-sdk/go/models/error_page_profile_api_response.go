// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ErrorPageProfileAPIResponse error page profile Api response
// swagger:model ErrorPageProfileApiResponse
type ErrorPageProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ErrorPageProfile `json:"results,omitempty"`
}
