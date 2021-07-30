// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LogAgentTCPClientEventDetail log agent TCP client event detail
// swagger:model LogAgentTCPClientEventDetail
type LogAgentTCPClientEventDetail struct {

	//  Field introduced in 20.1.3.
	ErrorCode *string `json:"error_code,omitempty"`

	//  Field introduced in 20.1.3.
	ErrorReason *string `json:"error_reason,omitempty"`

	//  Field introduced in 20.1.3.
	Host *string `json:"host,omitempty"`

	//  Field introduced in 20.1.3.
	Port *string `json:"port,omitempty"`
}
