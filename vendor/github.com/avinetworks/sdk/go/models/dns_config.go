package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSConfig DNS config
// swagger:model DNSConfig
type DNSConfig struct {

	// GSLB subdomain used for GSLB service FQDN match and placement. .
	// Required: true
	DomainName *string `json:"domain_name"`
}
