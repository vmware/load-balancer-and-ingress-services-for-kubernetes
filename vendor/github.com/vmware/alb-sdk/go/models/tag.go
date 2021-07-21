// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Tag tag
// swagger:model Tag
type Tag struct {

	//  Enum options - AVI_DEFINED, USER_DEFINED, VCENTER_DEFINED.
	Type *string `json:"type,omitempty"`

	// value of Tag.
	// Required: true
	Value *string `json:"value"`
}
