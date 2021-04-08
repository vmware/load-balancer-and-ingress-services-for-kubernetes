package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafApplicationSignatureAppVersion waf application signature app version
// swagger:model WafApplicationSignatureAppVersion
type WafApplicationSignatureAppVersion struct {

	// Name of an application in the rule set. Field introduced in 20.1.1.
	// Read Only: true
	Application *string `json:"application,omitempty"`

	// The last version of the rule set when the rules corresponding to the application changed. Field introduced in 20.1.1.
	// Read Only: true
	LastChangedRulesetVersion *string `json:"last_changed_ruleset_version,omitempty"`

	// The number of rules available for this application. Field introduced in 20.1.3.
	// Read Only: true
	NumberOfRules *int32 `json:"number_of_rules,omitempty"`
}
