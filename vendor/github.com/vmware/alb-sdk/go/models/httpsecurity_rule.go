// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpsecurityRule httpsecurity rule
// swagger:model HTTPSecurityRule
type HttpsecurityRule struct {

	// Action to be performed upon successful matching. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Action *HttpsecurityAction `json:"action,omitempty"`

	// Enable or disable the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Enable *bool `json:"enable"`

	// Index of the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	// Log HTTP request upon rule match. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Log *bool `json:"log,omitempty"`

	// Add match criteria to the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Match *MatchTarget `json:"match,omitempty"`

	// Name of the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`
}
