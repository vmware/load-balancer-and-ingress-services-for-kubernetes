// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CSRFRule c s r f rule
// swagger:model CSRFRule
type CSRFRule struct {

	// CSRF Action to be applied for matched target. Enum options - VERIFY_CSRF_TOKEN, VERIFY_ORIGIN, VERIFY_ORIGIN_AND_CSRF_TOKEN, BYPASS_CSRF. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Action *string `json:"action,omitempty"`

	// Enable or deactivate the rule. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`

	// Rules are processed in order of this index field. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	// Match criteria for requests to apply CSRF Action. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Match *MatchTarget `json:"match"`

	// A name describing the rule in a short form. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`
}
