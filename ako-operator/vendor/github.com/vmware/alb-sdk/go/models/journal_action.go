// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JournalAction journal action
// swagger:model JournalAction
type JournalAction struct {

	// Details of the process for each object type. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Objects []*JournalObject `json:"objects,omitempty"`

	// Migrated version. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Version *string `json:"version,omitempty"`
}
