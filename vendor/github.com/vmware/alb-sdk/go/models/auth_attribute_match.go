// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthAttributeMatch auth attribute match
// swagger:model AuthAttributeMatch
type AuthAttributeMatch struct {

	// Attribute name whose values will be looked up in the access lists. Field introduced in 18.2.5.
	// Required: true
	AttributeName *string `json:"attribute_name"`

	// Attribute Values used to determine access when authentication applies. Field introduced in 18.2.5. Allowed in Basic edition, Essentials edition, Enterprise edition.
	// Required: true
	AttributeValueList *StringMatch `json:"attribute_value_list"`
}
