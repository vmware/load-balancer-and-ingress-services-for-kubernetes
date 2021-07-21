// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigUserNotAuthrzByRule config user not authrz by rule
// swagger:model ConfigUserNotAuthrzByRule
type ConfigUserNotAuthrzByRule struct {

	// Comma separated list of policies assigned to the user. Field introduced in 18.2.7, 20.1.1.
	Policies *string `json:"policies,omitempty"`

	// assigned roles.
	Roles *string `json:"roles,omitempty"`

	// assigned tenants.
	Tenants *string `json:"tenants,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
