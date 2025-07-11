// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IcapLog icap log
// swagger:model IcapLog
type IcapLog struct {

	// Denotes whether the content was processed by ICAP server and an action was taken. Enum options - ICAP_DISABLED, ICAP_PASSED, ICAP_MODIFIED, ICAP_BLOCKED, ICAP_FAILED. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Action *string `json:"action,omitempty"`

	// Logs for the HTTP request's content sent to the ICAP server. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestLogs []*IcapRequestLog `json:"request_logs,omitempty"`
}
