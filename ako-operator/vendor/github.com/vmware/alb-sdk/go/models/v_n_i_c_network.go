// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VNICNetwork v n i c network
// swagger:model vNICNetwork
type VNICNetwork struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CtlrAlloc *bool `json:"ctlr_alloc,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IP *IPAddrPrefix `json:"ip"`

	//  Enum options - DHCP, STATIC, VIP, DOCKER_HOST, MODE_MANUAL. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Mode *string `json:"mode"`
}
