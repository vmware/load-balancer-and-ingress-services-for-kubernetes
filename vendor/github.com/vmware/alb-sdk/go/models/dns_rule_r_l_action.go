// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRuleRLAction Dns rule r l action
// swagger:model DnsRuleRLAction
type DNSRuleRLAction struct {

	// Type of action to be enforced upon hitting the rate limit. Enum options - DNS_RL_ACTION_NONE, DNS_RL_ACTION_DROP_REQ. Field introduced in 18.2.5.
	Type *string `json:"type,omitempty"`
}
