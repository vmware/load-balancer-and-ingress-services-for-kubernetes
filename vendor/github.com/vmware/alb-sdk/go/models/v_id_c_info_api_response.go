// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIDCInfoAPIResponse v ID c info Api response
// swagger:model VIDCInfoApiResponse
type VIDCInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIDCInfo `json:"results,omitempty"`
}
