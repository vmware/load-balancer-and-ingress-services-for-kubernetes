// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRuleActionResponse Dns rule action response
// swagger:model DnsRuleActionResponse
type DNSRuleActionResponse struct {

	// DNS response is authoritative. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Authoritative *bool `json:"authoritative,omitempty"`

	// DNS response code. Enum options - DNS_RCODE_NOERROR, DNS_RCODE_FORMERR, DNS_RCODE_SERVFAIL, DNS_RCODE_NXDOMAIN, DNS_RCODE_NOTIMP, DNS_RCODE_REFUSED, DNS_RCODE_YXDOMAIN, DNS_RCODE_YXRRSET, DNS_RCODE_NXRRSET, DNS_RCODE_NOTAUTH, DNS_RCODE_NOTZONE. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rcode *string `json:"rcode,omitempty"`

	// DNS resource record sets - (resource record set share the DNS domain name, type, and class). Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceRecordSets []*DNSRuleDNSRrSet `json:"resource_record_sets,omitempty"`

	// DNS response is truncated. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Truncation *bool `json:"truncation,omitempty"`
}
