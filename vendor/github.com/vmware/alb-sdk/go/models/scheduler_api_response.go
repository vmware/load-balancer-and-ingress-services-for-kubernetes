// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SchedulerAPIResponse scheduler Api response
// swagger:model SchedulerApiResponse
type SchedulerAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*Scheduler `json:"results,omitempty"`
}
