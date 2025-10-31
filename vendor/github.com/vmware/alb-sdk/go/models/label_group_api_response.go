// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LabelGroupAPIResponse label group Api response
// swagger:model LabelGroupApiResponse
type LabelGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*LabelGroup `json:"results,omitempty"`
}
