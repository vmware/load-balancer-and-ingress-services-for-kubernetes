// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthorizationRule authorization rule
// swagger:model AuthorizationRule
type AuthorizationRule struct {

	// Authorization action when rule is matched. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Action *AuthorizationAction `json:"action"`

	// Enable or disable the rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Enable *bool `json:"enable"`

	// Index of the Authorization Policy rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	// Authorization match criteria for the rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Match *AuthorizationMatch `json:"match"`

	// Name of the rule. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`
}
