// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DestinationPortAddr destination port addr
// swagger:model DestinationPortAddr
type DestinationPortAddr struct {

	// TCP/UDP port range end (inclusive). Allowed values are 1-65535. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DstPortEnd *uint32 `json:"dst_port_end,omitempty"`

	// TCP/UDP port range start (inclusive). Allowed values are 1-65535. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DstPortStart *uint32 `json:"dst_port_start,omitempty"`

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchOperation *string `json:"match_operation,omitempty"`
}
