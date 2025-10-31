// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DynamicDNSRecord dynamic Dns record
// swagger:model DynamicDnsRecord
type DynamicDNSRecord struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Specifies the algorithm to pick the IP address(es) to be returned,when multiple entries are configured. This does not apply if num_records_in_response is 0. Default is round-robin. Enum options - DNS_RECORD_RESPONSE_ROUND_ROBIN, DNS_RECORD_RESPONSE_CONSISTENT_HASH. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Algorithm *string `json:"algorithm,omitempty"`

	// Canonical name in CNAME record. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Cname *DNSCnameRdata `json:"cname,omitempty"`

	// Configured FQDNs are delegated domains (i.e. they represent a zone cut). Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Delegated *bool `json:"delegated,omitempty"`

	// Details of DNS record. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// UUID of the DNS VS. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSVsUUID *string `json:"dns_vs_uuid,omitempty"`

	// Fully Qualified Domain Name. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Fqdn *string `json:"fqdn,omitempty"`

	// IPv6 address in AAAA record. Field introduced in 20.1.3. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ip6Address []*DNSAAAARdata `json:"ip6_address,omitempty"`

	// IP address in A record. Field introduced in 20.1.3. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPAddress []*DNSARdata `json:"ip_address,omitempty"`

	// Internal metadata for the DNS record. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Metadata *string `json:"metadata,omitempty"`

	// MX record. Field introduced in 20.1.3. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MxRecords []*DNSMxRdata `json:"mx_records,omitempty"`

	// DynamicDnsRecord name, needed for a top level uuid protobuf, for display in shell. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Name Server information in NS record. Field introduced in 20.1.3. Maximum of 13 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ns []*DNSNsRdata `json:"ns,omitempty"`

	// Specifies the number of records returned by the DNS service.Enter 0 to return all records. Default is 0. Allowed values are 0-20. Special values are 0- Return all records. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumRecordsInResponse *uint32 `json:"num_records_in_response,omitempty"`

	// Service locator info in SRV record. Field introduced in 20.1.3. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceLocators []*DNSSrvRdata `json:"service_locators,omitempty"`

	// tenant_uuid from Dns VS's tenant_uuid. It is a reference to an object of type Tenant. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Time To Live for this DNS record. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TTL *uint32 `json:"ttl,omitempty"`

	// Text record. Field introduced in 20.1.3. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TxtRecords []*DNSTxtRdata `json:"txt_records,omitempty"`

	// DNS record type. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the dns record. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Enable wild-card match of fqdn  if an exact match is not found in the DNS table, the longest match is chosen by wild-carding the fqdn in the DNS request. Default is false. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	WildcardMatch *bool `json:"wildcard_match,omitempty"`
}
