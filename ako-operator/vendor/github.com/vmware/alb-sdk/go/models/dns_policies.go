// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSPolicies Dns policies
// swagger:model DnsPolicies
type DNSPolicies struct {

	// UUID of the dns policy. It is a reference to an object of type DnsPolicy. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DNSPolicyRef *string `json:"dns_policy_ref"`

	// Index of the dns policy. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`
}
