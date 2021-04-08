package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSResourceRecord Dns resource record
// swagger:model DnsResourceRecord
type DNSResourceRecord struct {

	// IPv6 address of the requested FQDN. Field introduced in 18.1.1.
	Addr6IPStr *string `json:"addr6_ip_str,omitempty"`

	// IPv4 address of the requested FQDN.
	AddrIP *int32 `json:"addr_ip,omitempty"`

	// Canonical (real) name of the requested FQDN.
	Cname *string `json:"cname,omitempty"`

	// Class of the data in the resource record.
	Dclass *int32 `json:"dclass,omitempty"`

	// Geo Location of Member. Field introduced in 17.1.1.
	Location *GeoLocation `json:"location,omitempty"`

	// Fully qualified domain name of a mail server in the MX record. Field introduced in 18.2.9, 20.1.1.
	MailServer *string `json:"mail_server,omitempty"`

	// Domain name of the resource record.
	Name *string `json:"name,omitempty"`

	// Domain name of the name server that is authoritative for the requested FQDN.
	Nsname *string `json:"nsname,omitempty"`

	// Service port.
	Port *int32 `json:"port,omitempty"`

	// The priority field identifies which mail server should be preferred. Field introduced in 18.2.9, 20.1.1.
	Priority *int32 `json:"priority,omitempty"`

	// Site controller cluster name - applicable only for Avi VS GSLB member.
	SiteName *string `json:"site_name,omitempty"`

	// Text resource record. Field introduced in 18.2.9, 20.1.1.
	TextRdata *string `json:"text_rdata,omitempty"`

	// Number of seconds the resource record can be cached.
	// Required: true
	TTL *int32 `json:"ttl"`

	// Type of resource record. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY.
	// Required: true
	Type *string `json:"type"`

	// Virtual Service name - applicable only for Avi VS GSLB member.
	VsName *string `json:"vs_name,omitempty"`
}
