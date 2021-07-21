// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRuleActionAllowDrop Dns rule action allow drop
// swagger:model DnsRuleActionAllowDrop
type DNSRuleActionAllowDrop struct {

	// Allow the DNS query. Field introduced in 17.1.1.
	Allow *bool `json:"allow,omitempty"`

	// Reset the TCP connection of the DNS query, if allow is set to false to drop the query. Field introduced in 17.1.1.
	ResetConn *bool `json:"reset_conn,omitempty"`
}
