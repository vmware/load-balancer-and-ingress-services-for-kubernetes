package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRuleMatchTarget Dns rule match target
// swagger:model DnsRuleMatchTarget
type DNSRuleMatchTarget struct {

	// IP addresses to match against client IP. From 17.1.6 release onwards, IP addresses needs to be configured in the client_ip_address field of this message. Field deprecated in 17.1.6,17.2.2. Field introduced in 17.1.1.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// IP addresses to match against client IP or the EDNS client subnet IP. Field introduced in 17.1.6,17.2.2.
	ClientIPAddress *DNSClientIPMatch `json:"client_ip_address,omitempty"`

	// Geographical location attribute to match against that of the client IP. Field introduced in 17.1.5.
	GeoLocation *DNSGeoLocationMatch `json:"geo_location,omitempty"`

	// DNS transport protocol match. Field introduced in 17.1.1.
	Protocol *DNSTransportProtocolMatch `json:"protocol,omitempty"`

	// Domain names to match against query name. Field introduced in 17.1.1.
	QueryName *DNSQueryNameMatch `json:"query_name,omitempty"`

	// DNS query types to match against request query type. Field introduced in 17.1.1.
	QueryType *DNSQueryTypeMatch `json:"query_type,omitempty"`
}
