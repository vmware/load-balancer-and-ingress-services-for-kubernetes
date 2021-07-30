// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MesosAttribute mesos attribute
// swagger:model MesosAttribute
type MesosAttribute struct {

	// Attribute to match.
	// Required: true
	Attribute *string `json:"attribute"`

	// Attribute value. If not set, match any value.
	Value *string `json:"value,omitempty"`
}
