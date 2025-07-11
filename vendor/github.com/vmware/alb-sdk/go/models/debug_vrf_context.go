// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugVrfContext debug vrf context
// swagger:model DebugVrfContext
type DebugVrfContext struct {

	// Vrf config command buffer process interval. Allowed values are 1-4. Field introduced in 17.2.13,18.1.5,18.2.1. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CommandBufferInterval *uint32 `json:"command_buffer_interval,omitempty"`

	// Vrf config command buffer size. Allowed values are 1-32768. Field introduced in 17.2.13,18.1.5,18.2.1. Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CommandBufferSize *uint32 `json:"command_buffer_size,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Flags []*DebugVrf `json:"flags,omitempty"`
}
