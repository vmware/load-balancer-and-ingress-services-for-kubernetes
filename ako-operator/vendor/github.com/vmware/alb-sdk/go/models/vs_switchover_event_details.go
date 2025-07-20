// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsSwitchoverEventDetails vs switchover event details
// swagger:model VsSwitchoverEventDetails
type VsSwitchoverEventDetails struct {

	// Error messages associated with this Event. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorMessage *string `json:"error_message,omitempty"`

	// VIP IPv4 address. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IP *string `json:"ip,omitempty"`

	// VIP IPv6 address. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ip6 *string `json:"ip6,omitempty"`

	// Status of Event. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RPCStatus *uint64 `json:"rpc_status,omitempty"`

	// List of ServiceEngine assigned to this Virtual Service. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeAssigned []*VipSeAssigned `json:"se_assigned,omitempty"`

	// Resources requested/assigned to this Virtual Service. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeRequested *VirtualServiceResource `json:"se_requested,omitempty"`

	// Virtual Service UUID. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
