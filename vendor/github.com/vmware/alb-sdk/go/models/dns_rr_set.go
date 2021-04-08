package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRrSet Dns rr set
// swagger:model DnsRrSet
type DNSRrSet struct {

	// Canonical name in CNAME record. Field introduced in 17.2.12, 18.1.2.
	Cname *DNSCnameRdata `json:"cname,omitempty"`

	// Fully Qualified Domain Name. Field introduced in 17.2.12, 18.1.2.
	// Required: true
	Fqdn *string `json:"fqdn"`

	// IPv6 address in AAAA record. Field introduced in 18.1.2.
	Ip6Addresses []*DNSAAAARdata `json:"ip6_addresses,omitempty"`

	// IP address in A record. Field introduced in 17.2.12, 18.1.2.
	IPAddresses []*DNSARdata `json:"ip_addresses,omitempty"`

	// Name Server information in NS record. Field introduced in 17.2.12, 18.1.2.
	Nses []*DNSNsRdata `json:"nses,omitempty"`

	// Time To Live for this DNS record. Allowed values are 0-2147483647. Field introduced in 17.2.12, 18.1.2.
	// Required: true
	TTL *int32 `json:"ttl"`

	// DNS record type. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY. Field introduced in 17.2.12, 18.1.2.
	// Required: true
	Type *string `json:"type"`
}
