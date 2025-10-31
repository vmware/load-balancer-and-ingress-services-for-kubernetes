// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LogAgentTCPClientEventDetail log agent TCP client event detail
// swagger:model LogAgentTCPClientEventDetail
type LogAgentTCPClientEventDetail struct {

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorCode *string `json:"error_code,omitempty"`

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorReason *string `json:"error_reason,omitempty"`

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Host *string `json:"host,omitempty"`

	//  Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Port *string `json:"port,omitempty"`
}
