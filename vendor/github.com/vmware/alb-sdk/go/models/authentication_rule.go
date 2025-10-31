// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthenticationRule authentication rule
// swagger:model AuthenticationRule
type AuthenticationRule struct {

	// Enable or disable authentication for matched targets. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Action *AuthenticationAction `json:"action,omitempty"`

	// Enable or disable the rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Enable *bool `json:"enable"`

	// Index of the rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	// Add match criteria to the rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Match *AuthenticationMatch `json:"match,omitempty"`

	// Name of the rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`
}
