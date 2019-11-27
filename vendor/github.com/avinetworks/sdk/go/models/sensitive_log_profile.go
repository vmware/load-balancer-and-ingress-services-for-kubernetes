package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SensitiveLogProfile sensitive log profile
// swagger:model SensitiveLogProfile
type SensitiveLogProfile struct {

	// Match sensitive header fields in HTTP application log. Field introduced in 17.2.10, 18.1.2.
	HeaderFieldRules []*SensitiveFieldRule `json:"header_field_rules,omitempty"`

	// Match sensitive WAF log fields in HTTP application log. Field introduced in 17.2.13, 18.1.3.
	WafFieldRules []*SensitiveFieldRule `json:"waf_field_rules,omitempty"`
}
