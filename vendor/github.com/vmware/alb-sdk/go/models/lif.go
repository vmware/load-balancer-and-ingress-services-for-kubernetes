// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Lif lif
// swagger:model Lif
type Lif struct {

	// Placeholder for description of property cifs of obj type Lif field type str  type object
	Cifs []*Cif `json:"cifs,omitempty"`

	// lif of Lif.
	Lif *string `json:"lif,omitempty"`

	// lif_label of Lif.
	LifLabel *string `json:"lif_label,omitempty"`

	// subnet of Lif.
	Subnet *string `json:"subnet,omitempty"`
}
