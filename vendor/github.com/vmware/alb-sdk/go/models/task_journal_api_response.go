// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TaskJournalAPIResponse task journal Api response
// swagger:model TaskJournalApiResponse
type TaskJournalAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TaskJournal `json:"results,omitempty"`
}
