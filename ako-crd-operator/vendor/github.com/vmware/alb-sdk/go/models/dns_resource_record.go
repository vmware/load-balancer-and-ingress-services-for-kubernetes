// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSResourceRecord Dns resource record
// swagger:model DnsResourceRecord
type DNSResourceRecord struct {

	// IPv6 address of the requested FQDN. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Addr6IPStr *string `json:"addr6_ip_str,omitempty"`

	// IPv4 address of the requested FQDN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AddrIP uint32 `json:"addr_ip,omitempty"`

	// Canonical (real) name of the requested FQDN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cname *string `json:"cname,omitempty"`

	// Class of the data in the resource record. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Dclass uint32 `json:"dclass,omitempty"`

	// Geo Location of Member. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Location *GeoLocation `json:"location,omitempty"`

	// Fully qualified domain name of a mail server in the MX record. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MailServer *string `json:"mail_server,omitempty"`

	// Domain name of the resource record. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Domain name of the name server that is authoritative for the requested FQDN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Nsname *string `json:"nsname,omitempty"`

	// Service port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port uint32 `json:"port,omitempty"`

	// The priority field identifies which mail server should be preferred. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Priority uint32 `json:"priority,omitempty"`

	// Site controller cluster name - applicable only for Avi VS GSLB member. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteName *string `json:"site_name,omitempty"`

	// Text resource record. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TextRdata *string `json:"text_rdata,omitempty"`

	// Number of seconds the resource record can be cached. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	TTL *uint32 `json:"ttl"`

	// Type of resource record. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// Virtual Service name - applicable only for Avi VS GSLB member. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsName *string `json:"vs_name,omitempty"`
}
