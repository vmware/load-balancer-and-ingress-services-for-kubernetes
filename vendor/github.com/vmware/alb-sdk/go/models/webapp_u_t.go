// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WebappUT webapp u t
// swagger:model WebappUT
type WebappUT struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Optional message for nested f_mandatory test cases defined at level1. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MandatoryTest *L1FMandatoryTestCase `json:"mandatory_test,omitempty"`

	// Repeated message for nested f_mandatory test cases-level1. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MandatoryTests []*L1FMandatoryTestCase `json:"mandatory_tests,omitempty"`

	// Name of the WebappUT object-level0. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Optional message for nested f_sensitive test cases defined at level1. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveTest *L1FSensitiveTestCase `json:"sensitive_test,omitempty"`

	// Repeated message for nested f_sensitive test cases-level1. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveTests []*L1FSensitiveTestCase `json:"sensitive_tests,omitempty"`

	// Optional message for nested  max *string length test cases. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StringLengthTest *L1StringLengthTestCase `json:"string_length_test,omitempty"`

	// Repeated message for nested  max *string length test cases. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StringLengthTests []*L1StringLengthTestCase `json:"string_length_tests,omitempty"`

	// Tenant of the WebappUT object-level0. It is a reference to an object of type Tenant. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// The *string for sensitive (secret) field.  object-level0. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TestSensitiveString *string `json:"test_sensitive_string,omitempty"`

	// The maximum *string length. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TestString *string `json:"test_string,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the WebappUT object-level0. Field introduced in 21.1.5, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
