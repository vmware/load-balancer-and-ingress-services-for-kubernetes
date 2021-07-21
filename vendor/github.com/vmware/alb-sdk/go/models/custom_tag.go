// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CustomTag custom tag
// swagger:model CustomTag
type CustomTag struct {

	// tag_key of CustomTag.
	// Required: true
	TagKey *string `json:"tag_key"`

	// tag_val of CustomTag.
	TagVal *string `json:"tag_val,omitempty"`
}
