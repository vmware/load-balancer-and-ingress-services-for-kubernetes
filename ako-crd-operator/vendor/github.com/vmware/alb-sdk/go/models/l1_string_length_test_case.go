// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L1StringLengthTestCase l1 *string length test case
// swagger:model L1StringLengthTestCase
type L1StringLengthTestCase struct {

	// String length message for nested *string length test cases. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StringLengthMessage *L2StringLengthTestCase `json:"string_length_message,omitempty"`

	// Repeated *string length message for nested *string length test cases. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StringLengthMessages []*L2StringLengthTestCase `json:"string_length_messages,omitempty"`

	// String field for nested *string length test cases. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TestString *string `json:"test_string,omitempty"`

	// Repeated field for nested *string length test cases. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TestStrings []string `json:"test_strings,omitempty"`
}
