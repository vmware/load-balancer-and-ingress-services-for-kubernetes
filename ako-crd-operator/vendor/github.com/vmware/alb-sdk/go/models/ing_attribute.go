// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IngAttribute ing attribute
// swagger:model IngAttribute
type IngAttribute struct {

	// Attribute to match. Field introduced in 17.2.15, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Attribute *string `json:"attribute,omitempty"`

	// Attribute value. If not set, match any value. Field introduced in 17.2.15, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
