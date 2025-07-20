// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigUserNotAuthrzByRule config user not authrz by rule
// swagger:model ConfigUserNotAuthrzByRule
type ConfigUserNotAuthrzByRule struct {

	// Comma separated list of policies assigned to the user. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Policies *string `json:"policies,omitempty"`

	// assigned roles. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Roles *string `json:"roles,omitempty"`

	// assigned tenants. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tenants *string `json:"tenants,omitempty"`

	// Request user. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`
}
