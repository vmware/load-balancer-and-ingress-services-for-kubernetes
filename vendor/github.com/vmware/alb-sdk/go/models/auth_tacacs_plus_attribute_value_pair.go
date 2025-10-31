// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthTacacsPlusAttributeValuePair auth tacacs plus attribute value pair
// swagger:model AuthTacacsPlusAttributeValuePair
type AuthTacacsPlusAttributeValuePair struct {

	// mandatory. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mandatory *bool `json:"mandatory,omitempty"`

	// attribute name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// attribute value. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
