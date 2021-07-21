// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MemberInterface member interface
// swagger:model MemberInterface
type MemberInterface struct {

	// Placeholder for description of property active of obj type MemberInterface field type str  type boolean
	Active *bool `json:"active,omitempty"`

	// if_name of MemberInterface.
	// Required: true
	IfName *string `json:"if_name"`

	//  Field introduced in 17.1.5.
	MacAddress *string `json:"mac_address,omitempty"`
}
