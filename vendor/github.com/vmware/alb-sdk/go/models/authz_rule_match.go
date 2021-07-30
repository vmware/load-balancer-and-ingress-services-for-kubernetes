// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthzRuleMatch authz rule match
// swagger:model AuthzRuleMatch
type AuthzRuleMatch struct {

	// Name of the executed Authorization rule Action. Enum options - ALLOW_ACCESS, CLOSE_CONNECTION, HTTP_LOCAL_RESPONSE. Field introduced in 20.1.3.
	RuleAction *string `json:"rule_action,omitempty"`

	// Name of the matched Authorization rule. Field introduced in 20.1.3.
	RuleName *string `json:"rule_name,omitempty"`
}
