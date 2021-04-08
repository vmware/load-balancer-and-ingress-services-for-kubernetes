package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRuleActionPoolSwitching Dns rule action pool switching
// swagger:model DnsRuleActionPoolSwitching
type DNSRuleActionPoolSwitching struct {

	// Reference of the pool group to serve the passthrough DNS query which cannot be served locally. It is a reference to an object of type PoolGroup. Field introduced in 18.1.3, 17.2.12.
	PoolGroupRef *string `json:"pool_group_ref,omitempty"`

	// Reference of the pool to serve the passthrough DNS query which cannot be served locally. It is a reference to an object of type Pool. Field introduced in 18.1.3, 17.2.12.
	PoolRef *string `json:"pool_ref,omitempty"`
}
