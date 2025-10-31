// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L1FSensitiveTestCase l1 f sensitive test case
// swagger:model L1FSensitiveTestCase
type L1FSensitiveTestCase struct {

	// f_sensitive message for nested f_sensitive test cases-level2. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveMessage *L2FSensitiveTestCase `json:"sensitive_message,omitempty"`

	// Repeated f_sensitive_message for nested f_sensitive test cases-level2. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveMessages []*L2FSensitiveTestCase `json:"sensitive_messages,omitempty"`

	// f_sensitive *string field for nested f_sensitive test cases-level1. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveString *string `json:"sensitive_string,omitempty"`
}
