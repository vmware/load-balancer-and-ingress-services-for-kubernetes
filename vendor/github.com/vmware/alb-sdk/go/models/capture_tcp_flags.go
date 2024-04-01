// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CaptureTCPFlags capture TCP flags
// swagger:model CaptureTCPFlags
type CaptureTCPFlags struct {

	// Logical operation based filter criteria. Enum options - OR, AND. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FilterOp *string `json:"filter_op,omitempty"`

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchOperation *string `json:"match_operation,omitempty"`

	// TCP ACK flag filter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TCPAck *bool `json:"tcp_ack,omitempty"`

	// TCP FIN flag filter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TCPFin *bool `json:"tcp_fin,omitempty"`

	// TCP PUSH flag filter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TCPPush *bool `json:"tcp_push,omitempty"`

	// TCP RST flag filter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TCPRst *bool `json:"tcp_rst,omitempty"`

	// TCP SYN flag filter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TCPSyn *bool `json:"tcp_syn,omitempty"`
}
