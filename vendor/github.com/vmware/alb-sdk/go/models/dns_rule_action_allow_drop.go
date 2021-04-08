package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRuleActionAllowDrop Dns rule action allow drop
// swagger:model DnsRuleActionAllowDrop
type DNSRuleActionAllowDrop struct {

	// Allow the DNS query. Field introduced in 17.1.1.
	Allow *bool `json:"allow,omitempty"`

	// Reset the TCP connection of the DNS query, if allow is set to false to drop the query. Field introduced in 17.1.1.
	ResetConn *bool `json:"reset_conn,omitempty"`
}
