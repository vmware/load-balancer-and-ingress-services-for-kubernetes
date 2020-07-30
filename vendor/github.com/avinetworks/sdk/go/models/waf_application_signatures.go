package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafApplicationSignatures waf application signatures
// swagger:model WafApplicationSignatures
type WafApplicationSignatures struct {

	// The external provide for the rules. It is a reference to an object of type WafApplicationSignatureProvider. Field introduced in 20.1.1.
	// Required: true
	ProviderRef *string `json:"provider_ref"`

	// The active application specific rules. You can change attributes like enabled, waf mode and exclusions, but not the rules itself. To change the rules, you can change the tags or the rule provider. Field introduced in 20.1.1.
	Rules []*WafRule `json:"rules,omitempty"`

	// The version in use of the provided ruleset. Field introduced in 20.1.1.
	// Read Only: true
	RulesetVersion *string `json:"ruleset_version,omitempty"`

	// List of applications for which we use the rules from the WafApplicationSignatureProvider. Field introduced in 20.1.1.
	SelectedApplications []string `json:"selected_applications,omitempty"`
}
