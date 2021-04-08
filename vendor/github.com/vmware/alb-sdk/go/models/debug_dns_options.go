package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugDNSOptions debug Dns options
// swagger:model DebugDnsOptions
type DebugDNSOptions struct {

	// This field filters the FQDN for Dns debug. Field introduced in 18.2.1. Maximum of 1 items allowed.
	DomainName []string `json:"domain_name,omitempty"`

	// This field filters the Gslb service for Dns debug. Field introduced in 18.2.1. Maximum of 1 items allowed.
	GslbServiceName []string `json:"gslb_service_name,omitempty"`
}
