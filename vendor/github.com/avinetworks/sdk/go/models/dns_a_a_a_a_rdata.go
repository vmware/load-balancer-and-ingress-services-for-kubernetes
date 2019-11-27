package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSAAAARdata Dns a a a a rdata
// swagger:model DnsAAAARdata
type DNSAAAARdata struct {

	// IPv6 address for FQDN. Field introduced in 18.1.1.
	// Required: true
	Ip6Address *IPAddr `json:"ip6_address"`
}
