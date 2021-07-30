// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NatRule nat rule
// swagger:model NatRule
type NatRule struct {

	// Nat rule Action Information. Field introduced in 18.2.3.
	// Required: true
	Action *NatPolicyAction `json:"action"`

	// Creator name. Field introduced in 18.2.3.
	CreatedBy *string `json:"created_by,omitempty"`

	// Nat rule enable flag. Field introduced in 18.2.3.
	// Required: true
	Enable *bool `json:"enable"`

	// Nat rule Index. Field introduced in 18.2.3.
	// Required: true
	Index *int32 `json:"index"`

	// Nat rule Match Criteria. Field introduced in 18.2.3.
	// Required: true
	Match *NatMatchTarget `json:"match"`

	// Nat rule Name. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`
}
