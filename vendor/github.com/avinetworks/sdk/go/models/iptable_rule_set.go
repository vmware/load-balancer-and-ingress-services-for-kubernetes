package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IptableRuleSet iptable rule set
// swagger:model IptableRuleSet
type IptableRuleSet struct {

	// chain of IptableRuleSet.
	// Required: true
	Chain *string `json:"chain"`

	// Placeholder for description of property rules of obj type IptableRuleSet field type str  type object
	Rules []*IptableRule `json:"rules,omitempty"`

	// table of IptableRuleSet.
	// Required: true
	Table *string `json:"table"`
}
