// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthMatchGroupMembership auth match group membership
// swagger:model AuthMatchGroupMembership
type AuthMatchGroupMembership struct {

	// rule match criteria. Enum options - AUTH_MATCH_CONTAINS, AUTH_MATCH_DOES_NOT_CONTAIN, AUTH_MATCH_REGEX.
	// Required: true
	Criteria *string `json:"criteria"`

	// groups of AuthMatchGroupMembership.
	Groups []string `json:"groups,omitempty"`
}
