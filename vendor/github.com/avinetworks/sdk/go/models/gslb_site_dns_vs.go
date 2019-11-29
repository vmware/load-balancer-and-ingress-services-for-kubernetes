package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSiteDNSVs gslb site Dns vs
// swagger:model GslbSiteDnsVs
type GslbSiteDNSVs struct {

	// This field identifies the DNS VS uuid for this site. Field introduced in 17.2.3.
	// Required: true
	DNSVsUUID *string `json:"dns_vs_uuid"`

	// This field identifies the subdomains that are hosted on the DNS VS. GslbService(s) whose FQDNs map to one of the subdomains will be hosted on this DNS VS. If no subdomains are configured, then the default behavior is to host all the GslbServices on this DNS VS. Field introduced in 17.2.3.
	DomainNames []string `json:"domain_names,omitempty"`
}
