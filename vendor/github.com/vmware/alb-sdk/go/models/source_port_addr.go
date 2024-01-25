// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SourcePortAddr source port addr
// swagger:model SourcePortAddr
type SourcePortAddr struct {

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchOperation *string `json:"match_operation,omitempty"`

	// TCP/UDP port range end (inclusive). Allowed values are 1-65535. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SrcPortEnd uint32 `json:"src_port_end,omitempty"`

	// TCP/UDP port range start (inclusive). Allowed values are 1-65535. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SrcPortStart uint32 `json:"src_port_start,omitempty"`
}
