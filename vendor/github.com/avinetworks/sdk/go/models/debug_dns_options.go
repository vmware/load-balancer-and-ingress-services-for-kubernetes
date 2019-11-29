package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugDNSOptions debug Dns options
// swagger:model DebugDnsOptions
type DebugDNSOptions struct {

	// This field filters the FQDN for Dns debug. Field introduced in 18.2.1.
	DomainName []string `json:"domain_name,omitempty"`

	// This field filters the Gslb service for Dns debug. Field introduced in 18.2.1.
	GslbServiceName []string `json:"gslb_service_name,omitempty"`
}
