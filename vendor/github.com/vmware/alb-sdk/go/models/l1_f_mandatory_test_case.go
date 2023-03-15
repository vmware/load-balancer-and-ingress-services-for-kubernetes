// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L1FMandatoryTestCase l1 f mandatory test case
// swagger:model L1FMandatoryTestCase
type L1FMandatoryTestCase struct {

	// f_mandatory message for nested f_mandatory test cases-level2. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryMessage *L2FMandatoryTestCase `json:"mandatory_message"`

	// Repeated f_mandatory_message for nested f_mandatory test cases-level2. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryMessages []*L2FMandatoryTestCase `json:"mandatory_messages,omitempty"`

	// f_mandatory *string field for nested f_mandatory test cases-level1. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryString *string `json:"mandatory_string"`

	// Repeated f_mandatory *string field for nested f_mandatory test cases-level1. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MandatoryStrings []string `json:"mandatory_strings,omitempty"`
}
