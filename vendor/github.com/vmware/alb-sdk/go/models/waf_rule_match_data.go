package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafRuleMatchData waf rule match data
// swagger:model WafRuleMatchData
type WafRuleMatchData struct {

	// The match_element is an internal variable. It is not possible to add exclusions for this element. Field introduced in 17.2.4.
	IsInternal *bool `json:"is_internal,omitempty"`

	// Field from a transaction that matches the rule, for instance if the request parameter is password=foobar, then match_element is ARGS password. Field introduced in 17.2.1.
	MatchElement *string `json:"match_element,omitempty"`

	// Value for a field from a transaction that matches the rule, for instance if the request parameter is password=foobar, then match_value is foobar. Field introduced in 17.2.1.
	MatchValue *string `json:"match_value,omitempty"`
}
