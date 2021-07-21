// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthTacacsPlusAttributeValuePair auth tacacs plus attribute value pair
// swagger:model AuthTacacsPlusAttributeValuePair
type AuthTacacsPlusAttributeValuePair struct {

	// mandatory.
	Mandatory *bool `json:"mandatory,omitempty"`

	// attribute name.
	Name *string `json:"name,omitempty"`

	// attribute value.
	Value *string `json:"value,omitempty"`
}
