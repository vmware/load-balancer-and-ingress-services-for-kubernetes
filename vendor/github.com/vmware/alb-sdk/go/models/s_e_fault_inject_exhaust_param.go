// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SEFaultInjectExhaustParam s e fault inject exhaust param
// swagger:model SEFaultInjectExhaustParam
type SEFaultInjectExhaustParam struct {

	// Placeholder for description of property leak of obj type SEFaultInjectExhaustParam field type str  type boolean
	Leak *bool `json:"leak,omitempty"`

	// Number of num.
	// Required: true
	Num *int64 `json:"num"`
}
