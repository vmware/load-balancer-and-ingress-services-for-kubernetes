// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RoleMatchOperationMatchLabel role match operation match label
// swagger:model RoleMatchOperationMatchLabel
type RoleMatchOperationMatchLabel struct {

	// List of labels allowed for the tenant. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MatchLabel *RoleFilterMatchLabel `json:"match_label"`

	// Label match operation criteria. Enum options - ROLE_FILTER_EQUALS, ROLE_FILTER_DOES_NOT_EQUAL, ROLE_FILTER_GLOB_MATCH, ROLE_FILTER_GLOB_DOES_NOT_MATCH. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchOperation *string `json:"match_operation,omitempty"`
}
