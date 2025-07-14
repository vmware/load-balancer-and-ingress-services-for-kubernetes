// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafApplicationSignatures waf application signatures
// swagger:model WafApplicationSignatures
type WafApplicationSignatures struct {

	// The external provide for the rules. It is a reference to an object of type WafApplicationSignatureProvider. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ProviderRef *string `json:"provider_ref"`

	// A resolved version of the active application specific rules together with the overrides. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ResolvedRules []*WafRule `json:"resolved_rules,omitempty"`

	// Override attributes of application signature rules. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleOverrides []*WafRuleOverrides `json:"rule_overrides,omitempty"`

	// The version in use of the provided ruleset. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	RulesetVersion *string `json:"ruleset_version,omitempty"`

	// List of applications for which we use the rules from the WafApplicationSignatureProvider. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SelectedApplications []string `json:"selected_applications,omitempty"`
}
