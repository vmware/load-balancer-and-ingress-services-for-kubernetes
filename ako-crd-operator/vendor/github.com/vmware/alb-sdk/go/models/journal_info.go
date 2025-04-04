// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JournalInfo journal info
// swagger:model JournalInfo
type JournalInfo struct {

	// Details of run for each version. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Actions []*JournalAction `json:"actions,omitempty"`

	// Number of objects to be processed. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	TotalObjects *uint32 `json:"total_objects"`

	// List of versions to be migrated. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Versions []string `json:"versions,omitempty"`
}
