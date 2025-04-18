// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigUserAuthrzByRule config user authrz by rule
// swagger:model ConfigUserAuthrzByRule
type ConfigUserAuthrzByRule struct {

	// Comma separated list of policies assigned to the user. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Policies *string `json:"policies,omitempty"`

	// assigned roles. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Roles *string `json:"roles,omitempty"`

	// matching rule string. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rule *string `json:"rule,omitempty"`

	// assigned tenants. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tenants *string `json:"tenants,omitempty"`

	// Request user. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`

	// assigned user account profile name. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Userprofile *string `json:"userprofile,omitempty"`
}
