// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SingleOptionalFieldMessage single optional field message
// swagger:model SingleOptionalFieldMessage
type SingleOptionalFieldMessage struct {

	// Optional *string field for nested f_mandatory test cases-level3. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OptionalString *string `json:"optional_string,omitempty"`
}
