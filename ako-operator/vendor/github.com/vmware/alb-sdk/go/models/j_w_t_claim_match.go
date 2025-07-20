// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JWTClaimMatch j w t claim match
// swagger:model JWTClaimMatch
type JWTClaimMatch struct {

	// Boolean value against which the claim is matched. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BoolMatch *bool `json:"bool_match,omitempty"`

	// Integer value against which the claim is matched. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IntMatch *uint32 `json:"int_match,omitempty"`

	// Specified Claim should be present in the JWT. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	IsMandatory *bool `json:"is_mandatory"`

	// JWT Claim name to be validated. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// String values against which the claim is matched. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StringMatch *StringMatch `json:"string_match,omitempty"`

	// Specifies the type of the Claim. Enum options - JWT_CLAIM_TYPE_BOOL, JWT_CLAIM_TYPE_INT, JWT_CLAIM_TYPE_STRING. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// Specifies whether to validate the Claim value. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Validate *bool `json:"validate"`
}
