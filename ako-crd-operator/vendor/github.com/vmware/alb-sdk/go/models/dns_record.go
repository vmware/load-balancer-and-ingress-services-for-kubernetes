// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRecord Dns record
// swagger:model DnsRecord
type DNSRecord struct {

	// Specifies the algorithm to pick the IP address(es) to be returned, when multiple entries are configured. This does not apply if num_records_in_response is 0. Default is round-robin. Enum options - DNS_RECORD_RESPONSE_ROUND_ROBIN, DNS_RECORD_RESPONSE_CONSISTENT_HASH. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Algorithm *string `json:"algorithm,omitempty"`

	// Canonical name in CNAME record. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cname *DNSCnameRdata `json:"cname,omitempty"`

	// Configured FQDNs are delegated domains (i.e. they represent a zone cut). Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Delegated *bool `json:"delegated,omitempty"`

	// Details of DNS record. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Fully Qualified Domain Name. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fqdn []string `json:"fqdn,omitempty"`

	// IPv6 address in AAAA record. Field introduced in 18.1.1. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ip6Address []*DNSAAAARdata `json:"ip6_address,omitempty"`

	// IP address in A record. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAddress []*DNSARdata `json:"ip_address,omitempty"`

	// Internal metadata for the DNS record. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Metadata *string `json:"metadata,omitempty"`

	// MX record. Field introduced in 18.2.9, 20.1.1. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MxRecords []*DNSMxRdata `json:"mx_records,omitempty"`

	// Name Server information in NS record. Field introduced in 17.1.1. Maximum of 13 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ns []*DNSNsRdata `json:"ns,omitempty"`

	// Specifies the number of records returned by the DNS service. Enter 0 to return all records. Default is 0. Allowed values are 0-20. Special values are 0- Return all records. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumRecordsInResponse uint32 `json:"num_records_in_response,omitempty"`

	// Service locator info in SRV record. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceLocator []*DNSSrvRdata `json:"service_locator,omitempty"`

	// Time To Live for this DNS record. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TTL uint32 `json:"ttl,omitempty"`

	// Text record. Field introduced in 18.2.9, 20.1.1. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TxtRecords []*DNSTxtRdata `json:"txt_records,omitempty"`

	// DNS record type. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// Enable wild-card match of fqdn  if an exact match is not found in the DNS table, the longest match is chosen by wild-carding the fqdn in the DNS request. Default is false. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WildcardMatch *bool `json:"wildcard_match,omitempty"`
}
