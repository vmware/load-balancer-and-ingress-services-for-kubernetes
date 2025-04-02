// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WebappUTAPIResponse webapp u t Api response
// swagger:model WebappUTApiResponse
type WebappUTAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*WebappUT `json:"results,omitempty"`
}
