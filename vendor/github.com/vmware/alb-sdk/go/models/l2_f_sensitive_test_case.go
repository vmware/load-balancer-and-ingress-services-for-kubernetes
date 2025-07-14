// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L2FSensitiveTestCase l2 f sensitive test case
// swagger:model L2FSensitiveTestCase
type L2FSensitiveTestCase struct {

	// f_sensitive message for nested f_sensitive test cases-level3. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveMessage *SingleOptionalSensitiveFieldMessage `json:"sensitive_message,omitempty"`

	// Repeated f_sensitive_message for nested f_sensitive test cases-level3. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveMessages []*SingleOptionalSensitiveFieldMessage `json:"sensitive_messages,omitempty"`

	// f_sensitive *string field for nested f_sensitive test cases-level2. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveString *string `json:"sensitive_string,omitempty"`
}
