// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRule Dns rule
// swagger:model DnsRule
type DNSRule struct {

	// Action to be performed upon successful matching. Field introduced in 17.1.1.
	Action *DNSRuleAction `json:"action,omitempty"`

	// Enable or disable the rule. Field introduced in 17.1.1.
	Enable *bool `json:"enable,omitempty"`

	// Index of the rule. Field introduced in 17.1.1.
	// Required: true
	Index *int32 `json:"index"`

	// Log DNS query upon rule match. Field introduced in 17.1.1.
	Log *bool `json:"log,omitempty"`

	// Add match criteria to the rule. Field introduced in 17.1.1.
	Match *DNSRuleMatchTarget `json:"match,omitempty"`

	// Name of the rule. Field introduced in 17.1.1.
	// Required: true
	Name *string `json:"name"`
}
