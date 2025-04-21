// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SingleOptionalSensitiveFieldMessage single optional sensitive field message
// swagger:model SingleOptionalSensitiveFieldMessage
type SingleOptionalSensitiveFieldMessage struct {

	// Optional *string field for nested f_mandatory test cases-level3. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OptionalSensitiveString *string `json:"optional_sensitive_string,omitempty"`
}
