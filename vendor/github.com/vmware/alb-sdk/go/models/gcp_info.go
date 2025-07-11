// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GcpInfo gcp info
// swagger:model GcpInfo
type GcpInfo struct {

	// Hostname of this SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hostname *string `json:"hostname,omitempty"`

	// Instance type of this SE. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MachineType *string `json:"machine_type,omitempty"`

	// Network this SE is assigned. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Network *string `json:"network"`

	// Project this SE belongs to. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Project *string `json:"project"`

	// Subnet assigned to this SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Subnet *string `json:"subnet,omitempty"`

	// Zone this SE is part of. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Zone *string `json:"zone"`
}
