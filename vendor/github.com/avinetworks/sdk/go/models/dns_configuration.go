package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSConfiguration DNS configuration
// swagger:model DNSConfiguration
type DNSConfiguration struct {

	// Search domain to use in DNS lookup.
	SearchDomain *string `json:"search_domain,omitempty"`

	// List of DNS Server IP addresses.
	ServerList []*IPAddr `json:"server_list,omitempty"`
}
