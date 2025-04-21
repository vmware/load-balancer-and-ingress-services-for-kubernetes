// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PortRange port range
// swagger:model PortRange
type PortRange struct {

	// TCP/UDP port range end (inclusive). Allowed values are 1-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	End *uint32 `json:"end"`

	// TCP/UDP port range start (inclusive). Allowed values are 1-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Start *uint32 `json:"start"`
}
