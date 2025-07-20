// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JournalObject journal object
// swagger:model JournalObject
type JournalObject struct {

	// Number of object caused a failure. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Failed *uint32 `json:"failed,omitempty"`

	// Name of the model. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Model *string `json:"model,omitempty"`

	// Number of object skipped. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Skipped *uint32 `json:"skipped,omitempty"`

	// Number of object for which processing is successful. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Success *uint32 `json:"success,omitempty"`
}
