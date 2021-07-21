// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HostAttributes host attributes
// swagger:model HostAttributes
type HostAttributes struct {

	// attr_key of HostAttributes.
	// Required: true
	AttrKey *string `json:"attr_key"`

	// attr_val of HostAttributes.
	AttrVal *string `json:"attr_val,omitempty"`
}
