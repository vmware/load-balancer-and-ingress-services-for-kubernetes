// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIPGNameInfoAPIResponse v IP g name info Api response
// swagger:model VIPGNameInfoApiResponse
type VIPGNameInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIPGNameInfo `json:"results,omitempty"`
}
