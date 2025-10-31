// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafApplicationSignatureAppVersion waf application signature app version
// swagger:model WafApplicationSignatureAppVersion
type WafApplicationSignatureAppVersion struct {

	// Name of an application in the rule set. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	Application *string `json:"application,omitempty"`

	// The last version of the rule set when the rules corresponding to the application changed. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	LastChangedRulesetVersion *string `json:"last_changed_ruleset_version,omitempty"`

	// The number of rules available for this application. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	NumberOfRules *uint32 `json:"number_of_rules,omitempty"`
}
