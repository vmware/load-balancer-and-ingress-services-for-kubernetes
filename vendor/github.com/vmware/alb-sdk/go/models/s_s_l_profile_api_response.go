// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLProfileAPIResponse s s l profile Api response
// swagger:model SSLProfileApiResponse
type SSLProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SSLProfile `json:"results,omitempty"`
}
