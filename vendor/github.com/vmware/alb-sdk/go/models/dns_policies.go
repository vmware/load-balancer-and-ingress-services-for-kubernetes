// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSPolicies Dns policies
// swagger:model DnsPolicies
type DNSPolicies struct {

	// UUID of the dns policy. It is a reference to an object of type DnsPolicy. Field introduced in 17.1.1.
	// Required: true
	DNSPolicyRef *string `json:"dns_policy_ref"`

	// Index of the dns policy. Field introduced in 17.1.1.
	// Required: true
	Index *int32 `json:"index"`
}
