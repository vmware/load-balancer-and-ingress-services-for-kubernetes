// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// KeyValue key value
// swagger:model KeyValue
type KeyValue struct {

	// Key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Key *string `json:"key"`

	// Value. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
