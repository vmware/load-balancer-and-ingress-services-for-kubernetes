package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyAllowlist waf policy allowlist
// swagger:model WafPolicyAllowlist
type WafPolicyAllowlist struct {

	// Rules to bypass WAF. Field introduced in 20.1.3. Maximum of 1024 items allowed.
	Rules []*WafPolicyAllowlistRule `json:"rules,omitempty"`
}
