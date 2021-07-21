// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// KeyValue key value
// swagger:model KeyValue
type KeyValue struct {

	// Key.
	// Required: true
	Key *string `json:"key"`

	// Value.
	Value *string `json:"value,omitempty"`
}
