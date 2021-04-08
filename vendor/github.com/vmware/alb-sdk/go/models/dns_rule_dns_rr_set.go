package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRuleDNSRrSet Dns rule Dns rr set
// swagger:model DnsRuleDnsRrSet
type DNSRuleDNSRrSet struct {

	// DNS resource record set - (records in the resource record set share the DNS domain name, type, and class). Field introduced in 17.2.12, 18.1.2.
	// Required: true
	ResourceRecordSet *DNSRrSet `json:"resource_record_set"`

	// DNS message section for the resource record set. Enum options - DNS_MESSAGE_SECTION_QUESTION, DNS_MESSAGE_SECTION_ANSWER, DNS_MESSAGE_SECTION_AUTHORITY, DNS_MESSAGE_SECTION_ADDITIONAL. Field introduced in 17.2.12, 18.1.2.
	Section *string `json:"section,omitempty"`
}
