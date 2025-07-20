// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterTagEventDetails vcenter tag event details
// swagger:model VcenterTagEventDetails
type VcenterTagEventDetails struct {

	// Cloud id. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CcID *string `json:"cc_id,omitempty"`

	// Failure reason. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorString *string `json:"error_string,omitempty"`

	// SEVM object id. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VMID *string `json:"vm_id,omitempty"`
}
