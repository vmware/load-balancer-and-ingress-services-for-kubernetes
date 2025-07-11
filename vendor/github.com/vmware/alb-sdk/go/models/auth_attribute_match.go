// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthAttributeMatch auth attribute match
// swagger:model AuthAttributeMatch
type AuthAttributeMatch struct {

	// Attribute name whose values will be looked up in the access lists. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AttributeName *string `json:"attribute_name"`

	// Attribute Values used to determine access when authentication applies. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	AttributeValueList *StringMatch `json:"attribute_value_list"`
}
