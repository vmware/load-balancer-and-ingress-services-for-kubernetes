package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSNsRdata Dns ns rdata
// swagger:model DnsNsRdata
type DNSNsRdata struct {

	// IPv6 address for Name Server. Field introduced in 18.1.1.
	Ip6Address *IPAddr `json:"ip6_address,omitempty"`

	// IP address for Name Server. Field introduced in 17.1.1.
	IPAddress *IPAddr `json:"ip_address,omitempty"`

	// Name Server name. Field introduced in 17.1.1.
	// Required: true
	Nsname *string `json:"nsname"`
}
