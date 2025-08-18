// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VSDataScriptSetAPIResponse v s data script set Api response
// swagger:model VSDataScriptSetApiResponse
type VSDataScriptSetAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VSDataScriptSet `json:"results,omitempty"`
}
