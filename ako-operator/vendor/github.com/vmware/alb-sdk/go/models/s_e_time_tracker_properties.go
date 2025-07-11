// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SETimeTrackerProperties s e time tracker properties
// swagger:model SETimeTrackerProperties
type SETimeTrackerProperties struct {

	// Audit queueing latency from proxy to dispatcher. Enum options - SE_TT_AUDIT_OFF, SE_TT_AUDIT_ON, SE_TT_AUDIT_ON_WITH_EVENT. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EgressAuditMode *string `json:"egress_audit_mode,omitempty"`

	// Maximum egress latency threshold between dispatcher and proxy. Field introduced in 22.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EgressThreshold *uint32 `json:"egress_threshold,omitempty"`

	// Window for cumulative event generation. Field introduced in 22.1.1. Unit is SECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EventGenWindow *uint64 `json:"event_gen_window,omitempty"`

	// Audit queueing latency from dispatcher to proxy. Enum options - SE_TT_AUDIT_OFF, SE_TT_AUDIT_ON, SE_TT_AUDIT_ON_WITH_EVENT. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IngressAuditMode *string `json:"ingress_audit_mode,omitempty"`

	// Maximum ingress latency threshold between dispatcher and proxy. Field introduced in 22.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IngressThreshold *uint32 `json:"ingress_threshold,omitempty"`
}
