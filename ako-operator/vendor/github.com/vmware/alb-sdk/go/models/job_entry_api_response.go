// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JobEntryAPIResponse job entry Api response
// swagger:model JobEntryApiResponse
type JobEntryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*JobEntry `json:"results,omitempty"`
}
