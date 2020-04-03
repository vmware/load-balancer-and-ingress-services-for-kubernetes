package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRuleRLAction Dns rule r l action
// swagger:model DnsRuleRLAction
type DNSRuleRLAction struct {

	// Type of action to be enforced upon hitting the rate limit. Enum options - DNS_RL_ACTION_NONE, DNS_RL_ACTION_DROP_REQ. Field introduced in 18.2.5.
	Type *string `json:"type,omitempty"`
}
