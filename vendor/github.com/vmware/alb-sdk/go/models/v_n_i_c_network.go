// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VNICNetwork v n i c network
// swagger:model vNICNetwork
type VNICNetwork struct {

	// Placeholder for description of property ctlr_alloc of obj type vNICNetwork field type str  type boolean
	CtlrAlloc *bool `json:"ctlr_alloc,omitempty"`

	// Placeholder for description of property ip of obj type vNICNetwork field type str  type object
	// Required: true
	IP *IPAddrPrefix `json:"ip"`

	//  Enum options - DHCP, STATIC, VIP, DOCKER_HOST.
	// Required: true
	Mode *string `json:"mode"`
}
