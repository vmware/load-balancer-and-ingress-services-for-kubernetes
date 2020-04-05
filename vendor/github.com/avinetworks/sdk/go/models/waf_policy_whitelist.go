package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyWhitelist waf policy whitelist
// swagger:model WafPolicyWhitelist
type WafPolicyWhitelist struct {

	// Rules to bypass WAF. Field introduced in 18.2.3.
	Rules []*WafPolicyWhitelistRule `json:"rules,omitempty"`
}
