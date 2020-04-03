package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NatPolicyAction nat policy action
// swagger:model NatPolicyAction
type NatPolicyAction struct {

	// Pool of IP Addresses used for Nat. Field introduced in 18.2.5.
	NatInfo []*NatAddrInfo `json:"nat_info,omitempty"`

	// Nat Action Type. Enum options - NAT_POLICY_ACTION_TYPE_DYNAMIC_IP_PORT. Field introduced in 18.2.5.
	// Required: true
	Type *string `json:"type"`
}
