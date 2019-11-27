package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSGeoLocationMatch Dns geo location match
// swagger:model DnsGeoLocationMatch
type DNSGeoLocationMatch struct {

	// Geographical location of the client IP to be used in the match. This location is of the format Country/State/City e.g. US/CA/Santa Clara. Field introduced in 17.1.5.
	GeolocationName *string `json:"geolocation_name,omitempty"`

	// Geolocation tag for the client IP. This could be any *string value for the client IP, e.g. client IPs from US East Coast geolocation would be tagged as 'East Coast'. Field introduced in 17.1.5.
	GeolocationTag *string `json:"geolocation_tag,omitempty"`

	// Criterion to use for matching the client IP's geographical location. Enum options - IS_IN, IS_NOT_IN. Field introduced in 17.1.5.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Use the IP address from the EDNS client subnet option, if available, to derive geo location of the DNS query. Field introduced in 17.1.5.
	UseEdnsClientSubnetIP *bool `json:"use_edns_client_subnet_ip,omitempty"`
}
