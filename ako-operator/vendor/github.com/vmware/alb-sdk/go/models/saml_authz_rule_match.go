// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlAuthzRuleMatch saml authz rule match
// swagger:model SamlAuthzRuleMatch
type SamlAuthzRuleMatch struct {

	// Name of the executed SAML Authorization rule Action. Enum options - ALLOW_ACCESS, CLOSE_CONNECTION, HTTP_LOCAL_RESPONSE. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthzMatchedRuleAction *string `json:"saml_authz_matched_rule_action,omitempty"`

	// Name of the matched SAML Authorization rule. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthzMatchedRuleName *string `json:"saml_authz_matched_rule_name,omitempty"`
}
