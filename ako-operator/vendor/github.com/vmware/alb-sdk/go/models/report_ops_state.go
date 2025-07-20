// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ReportOpsState report ops state
// swagger:model ReportOpsState
type ReportOpsState struct {

	// The last time the state changed. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Descriptive reason for the state-change. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// The system operation's current fsm-state. Enum options - SYSTEM_REPORT_STARTED, SYSTEM_REPORT_IN_PROGRESS, SYSTEM_REPORT_SUCCESS, SYSTEM_REPORT_WARNING, SYSTEM_REPORT_ERROR. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`
}
