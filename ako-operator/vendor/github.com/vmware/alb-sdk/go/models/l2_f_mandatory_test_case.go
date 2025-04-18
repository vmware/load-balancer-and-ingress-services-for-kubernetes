// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L2FMandatoryTestCase l2 f mandatory test case
// swagger:model L2FMandatoryTestCase
type L2FMandatoryTestCase struct {

	// f_mandatory message for nested f_mandatory test cases-level3. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryMessage *SingleOptionalFieldMessage `json:"mandatory_message"`

	// Repeated f_mandatory_message for nested f_mandatory test cases-level3. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryMessages []*SingleOptionalFieldMessage `json:"mandatory_messages,omitempty"`

	// f_mandatory *string field for nested f_mandatory test cases-level2. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryString *string `json:"mandatory_string"`

	// Repeated f_mandatory *string field for nested f_mandatory test cases-level2. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryStrings []string `json:"mandatory_strings,omitempty"`
}
