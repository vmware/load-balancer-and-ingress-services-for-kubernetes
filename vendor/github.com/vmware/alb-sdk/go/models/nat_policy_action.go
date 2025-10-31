// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NatPolicyAction nat policy action
// swagger:model NatPolicyAction
type NatPolicyAction struct {

	// Pool of IP Addresses used for Nat. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NatInfo []*NatAddrInfo `json:"nat_info,omitempty"`

	// Nat Action Type. Enum options - NAT_POLICY_ACTION_TYPE_DYNAMIC_IP_PORT, NAT_POLICY_ACTION_TYPE_DYNAMIC_IP_PRESERVE_PORT. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
