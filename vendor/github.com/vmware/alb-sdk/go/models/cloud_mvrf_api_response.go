// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudMvrfAPIResponse cloud mvrf Api response
// swagger:model CloudMvrfApiResponse
type CloudMvrfAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*CloudMvrf `json:"results,omitempty"`
}
