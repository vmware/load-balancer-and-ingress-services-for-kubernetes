// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSInfo Dns info
// swagger:model DnsInfo
type DNSInfo struct {

	// Specifies the algorithm to pick the IP address(es) to be returned, when multiple entries are configured. This does not apply if num_records_in_response is 0. Default is consistent hash. Enum options - DNS_RECORD_RESPONSE_ROUND_ROBIN, DNS_RECORD_RESPONSE_CONSISTENT_HASH. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Algorithm *string `json:"algorithm,omitempty"`

	// Canonical name in CNAME record. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cname *DNSCnameRdata `json:"cname,omitempty"`

	// Fully qualified domain name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fqdn *string `json:"fqdn,omitempty"`

	// Any metadata associated with this record. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	Metadata *string `json:"metadata,omitempty"`

	// Specifies the number of records returned for this FQDN. Enter 0 to return all records. Default is 0. Allowed values are 0-20. Special values are 0- Return all records. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumRecordsInResponse uint32 `json:"num_records_in_response,omitempty"`

	// Time to live for fqdn record. Default value is chosen from DNS profile for this cloud if no value provided. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TTL uint32 `json:"ttl,omitempty"`

	// DNS record type. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`
}
