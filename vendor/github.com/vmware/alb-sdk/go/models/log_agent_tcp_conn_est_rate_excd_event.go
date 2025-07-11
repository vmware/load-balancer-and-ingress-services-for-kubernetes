// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LogAgentTCPConnEstRateExcdEvent log agent TCP conn est rate excd event
// swagger:model LogAgentTCPConnEstRateExcdEvent
type LogAgentTCPConnEstRateExcdEvent struct {

	//  Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorCode *string `json:"error_code,omitempty"`

	//  Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorReason *string `json:"error_reason,omitempty"`

	//  Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Host *string `json:"host,omitempty"`

	//  Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Port *string `json:"port,omitempty"`
}
