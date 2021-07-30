// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NatPolicyAction nat policy action
// swagger:model NatPolicyAction
type NatPolicyAction struct {

	// Pool of IP Addresses used for Nat. Field introduced in 18.2.5.
	NatInfo []*NatAddrInfo `json:"nat_info,omitempty"`

	// Nat Action Type. Enum options - NAT_POLICY_ACTION_TYPE_DYNAMIC_IP_PORT. Field introduced in 18.2.5.
	// Required: true
	Type *string `json:"type"`
}
