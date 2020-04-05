package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRuleAction Dns rule action
// swagger:model DnsRuleAction
type DNSRuleAction struct {

	// Allow or drop the DNS query. Field introduced in 17.1.1.
	Allow *DNSRuleActionAllowDrop `json:"allow,omitempty"`

	// Rate limit the DNS requests. Field introduced in 18.2.5.
	DNSRateLimit *DNSRateProfile `json:"dns_rate_limit,omitempty"`

	// Select a specific GSLB site for the DNS query. This action should be used only when GSLB services have been configured for the DNS virtual service. Field introduced in 17.1.5.
	GslbSiteSelection *DNSRuleActionGslbSiteSelection `json:"gslb_site_selection,omitempty"`

	// Select a pool or pool group for the passthrough DNS query which cannot be served locally but could be served by upstream servers. Field introduced in 18.1.3, 17.2.12.
	PoolSwitching *DNSRuleActionPoolSwitching `json:"pool_switching,omitempty"`

	// Generate a response for the DNS query. Field introduced in 17.1.1.
	Response *DNSRuleActionResponse `json:"response,omitempty"`
}
