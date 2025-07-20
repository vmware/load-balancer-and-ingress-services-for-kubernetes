// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TimeTrackerProperties time tracker properties
// swagger:model TimeTrackerProperties
type TimeTrackerProperties struct {

	// Audit TCP connection establishment time on server-side. Enum options - TT_AUDIT_OFF, TT_AUDIT_ON, TT_AUDIT_ON_WITH_SIG. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BeConnEstAuditMode *string `json:"be_conn_est_audit_mode,omitempty"`

	// Maximum threshold for TCP connection establishment time on server-side. Field introduced in 22.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BeConnEstThreshold *uint32 `json:"be_conn_est_threshold,omitempty"`

	// Audit TCP connection establishment time on client-side. Enum options - TT_AUDIT_OFF, TT_AUDIT_ON, TT_AUDIT_ON_WITH_SIG. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FeConnEstAuditMode *string `json:"fe_conn_est_audit_mode,omitempty"`

	// Maximum threshold for TCP connection establishment time on client-side. Field introduced in 22.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FeConnEstThreshold *uint32 `json:"fe_conn_est_threshold,omitempty"`

	// Add significance if ingress latency from dispatcher to proxy is breached on any flow. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IngressSigLog *bool `json:"ingress_sig_log,omitempty"`
}
