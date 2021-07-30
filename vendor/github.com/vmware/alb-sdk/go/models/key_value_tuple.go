// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// KeyValueTuple key value tuple
// swagger:model KeyValueTuple
type KeyValueTuple struct {

	// Key. Field introduced in 20.1.3.
	// Required: true
	Key *string `json:"key"`

	// Value. Field introduced in 20.1.3.
	Value *string `json:"value,omitempty"`
}
