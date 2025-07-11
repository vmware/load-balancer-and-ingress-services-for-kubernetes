// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LatencyAuditProperties latency audit properties
// swagger:model LatencyAuditProperties
type LatencyAuditProperties struct {

	// Deprecated in 22.1.1. Enum options - LATENCY_AUDIT_OFF, LATENCY_AUDIT_ON, LATENCY_AUDIT_ON_WITH_SIG. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ConnEstAuditMode *string `json:"conn_est_audit_mode,omitempty"`

	// Deprecated in 22.1.1. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ConnEstThreshold *uint32 `json:"conn_est_threshold,omitempty"`

	// Deprecated in 22.1.1. Enum options - LATENCY_AUDIT_OFF, LATENCY_AUDIT_ON, LATENCY_AUDIT_ON_WITH_SIG. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LatencyAuditMode *string `json:"latency_audit_mode,omitempty"`

	// Deprecated in 22.1.1. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LatencyThreshold *uint32 `json:"latency_threshold,omitempty"`
}
