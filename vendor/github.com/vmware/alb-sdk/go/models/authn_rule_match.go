// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthnRuleMatch authn rule match
// swagger:model AuthnRuleMatch
type AuthnRuleMatch struct {

	// Name of the executed Authentication rule Action. Enum options - SKIP_AUTHENTICATION, USE_DEFAULT_AUTHENTICATION. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleAction *string `json:"rule_action,omitempty"`

	// Name of the matched Authentication rule. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleName *string `json:"rule_name,omitempty"`
}
