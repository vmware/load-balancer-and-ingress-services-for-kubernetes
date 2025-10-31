// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlAuthnRuleMatch saml authn rule match
// swagger:model SamlAuthnRuleMatch
type SamlAuthnRuleMatch struct {

	// Name of the executed SAML Authentication rule Action. Enum options - SKIP_AUTHENTICATION, USE_DEFAULT_AUTHENTICATION. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthnMatchedRuleAction *string `json:"saml_authn_matched_rule_action,omitempty"`

	// Name of the matched SAML Authentication rule. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthnMatchedRuleName *string `json:"saml_authn_matched_rule_name,omitempty"`
}
