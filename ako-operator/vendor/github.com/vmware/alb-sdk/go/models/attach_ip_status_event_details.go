// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AttachIPStatusEventDetails attach Ip status event details
// swagger:model AttachIpStatusEventDetails
type AttachIPStatusEventDetails struct {

	// Reason if Attach IP failed. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Name of the Service Engine. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// VIP ID. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VipID *string `json:"vip_id,omitempty"`

	// Name of the Virtual Service. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsName *string `json:"vs_name,omitempty"`
}
