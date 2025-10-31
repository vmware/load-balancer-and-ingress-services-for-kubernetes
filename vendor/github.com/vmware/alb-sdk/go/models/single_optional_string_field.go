// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SingleOptionalStringField single optional *string field
// swagger:model SingleOptionalStringField
type SingleOptionalStringField struct {

	// Optional *string field. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TestString *string `json:"test_string,omitempty"`
}
