// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkSecurityRule network security rule
// swagger:model NetworkSecurityRule
type NetworkSecurityRule struct {

	//  Enum options - NETWORK_SECURITY_POLICY_ACTION_TYPE_ALLOW, NETWORK_SECURITY_POLICY_ACTION_TYPE_DENY, NETWORK_SECURITY_POLICY_ACTION_TYPE_RATE_LIMIT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- NETWORK_SECURITY_POLICY_ACTION_TYPE_DENY), Basic edition(Allowed values- NETWORK_SECURITY_POLICY_ACTION_TYPE_DENY), Enterprise with Cloud Services edition.
	// Required: true
	Action *string `json:"action"`

	// Time in minutes after which rule will be deleted. Allowed values are 1-4294967295. Special values are 0- blocked for ever. Unit is MIN. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	Age *uint32 `json:"age,omitempty"`

	// Creator name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Enable *bool `json:"enable"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	//  Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	Log *bool `json:"log,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Match *NetworkSecurityMatchTarget `json:"match"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RlParam *NetworkSecurityPolicyActionRLParam `json:"rl_param,omitempty"`
}
